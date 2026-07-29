//! Explicit, redacting, bounded external-process execution.

use std::collections::BTreeMap;
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::thread;
use std::time::{Duration, Instant};

use crate::error::AinfraError;

/// Sanitized result of one explicit argument-vector invocation.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct ProcessResult {
    /// Exact non-secret argument vector.
    pub argv: Vec<String>,
    /// Process exit status, or `-1` when no numeric status is available.
    pub return_code: i32,
    /// Captured and redacted standard output.
    pub stdout: String,
    /// Captured and redacted standard error.
    pub stderr: String,
}

/// Complete safe subprocess request.
#[derive(Clone)]
pub struct ProcessRequest {
    /// Explicit executable and arguments; never interpreted by a shell.
    pub argv: Vec<String>,
    /// Controlled working directory.
    pub cwd: PathBuf,
    /// Complete child environment.
    pub environment: BTreeMap<String, String>,
    /// Values redacted from both output streams.
    pub secrets: Vec<String>,
}

/// Injectable execution boundary for lifecycle tests.
pub trait Runner {
    /// Execute one request or return a stable dependency failure.
    ///
    /// # Errors
    ///
    /// Returns `AINFRA-E500` when spawning, waiting, timing out, or collecting
    /// output fails.
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError>;
}

/// Real subprocess runner with a timeout and bounded retained output.
pub struct SubprocessRunner {
    timeout: Duration,
    output_limit: usize,
}

impl Default for SubprocessRunner {
    fn default() -> Self {
        Self {
            timeout: Duration::from_mins(10),
            output_limit: 1024 * 1024,
        }
    }
}

impl SubprocessRunner {
    /// Construct a runner with explicit execution limits.
    #[must_use]
    pub const fn new(timeout: Duration, output_limit: usize) -> Self {
        Self {
            timeout,
            output_limit,
        }
    }
}

impl Runner for SubprocessRunner {
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError> {
        let executable = request
            .argv
            .first()
            .ok_or_else(|| AinfraError::dependency("empty process argument vector"))?;
        let mut child = Command::new(executable)
            .args(&request.argv[1..])
            .current_dir(&request.cwd)
            .env_clear()
            .envs(&request.environment)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|error| {
                AinfraError::dependency(format!(
                    "required executable could not start: {executable}: {error}"
                ))
            })?;
        let stdout = child
            .stdout
            .take()
            .ok_or_else(|| AinfraError::dependency("cannot capture process stdout"))?;
        let stderr = child
            .stderr
            .take()
            .ok_or_else(|| AinfraError::dependency("cannot capture process stderr"))?;
        let limit = self.output_limit;
        let stdout_thread = thread::spawn(move || read_bounded(stdout, limit));
        let stderr_thread = thread::spawn(move || read_bounded(stderr, limit));
        let started = Instant::now();
        let status = loop {
            if let Some(status) = child
                .try_wait()
                .map_err(|error| AinfraError::dependency(error.to_string()))?
            {
                break status;
            }
            if started.elapsed() >= self.timeout {
                let _ = child.kill();
                let _ = child.wait();
                return Err(AinfraError::dependency(format!(
                    "underlying command timed out: {executable}"
                )));
            }
            thread::sleep(Duration::from_millis(10));
        };
        let stdout = join_output(stdout_thread, "stdout")?;
        let stderr = join_output(stderr_thread, "stderr")?;
        Ok(ProcessResult {
            argv: request.argv.clone(),
            return_code: status.code().unwrap_or(-1),
            stdout: redact(&String::from_utf8_lossy(&stdout), &request.secrets),
            stderr: redact(&String::from_utf8_lossy(&stderr), &request.secrets),
        })
    }
}

/// Build the minimal inherited environment used for child tools.
#[must_use]
pub fn child_environment() -> BTreeMap<String, String> {
    ["PATH", "HOME", "LANG", "LC_ALL", "TMPDIR"]
        .into_iter()
        .filter_map(|name| {
            std::env::var(name)
                .ok()
                .map(|value| (name.to_owned(), value))
        })
        .collect()
}

/// Replace every known non-empty secret with a stable marker.
#[must_use]
pub fn redact(value: &str, secrets: &[String]) -> String {
    let mut output = secrets
        .iter()
        .filter(|secret| !secret.is_empty())
        .fold(value.to_owned(), |output, secret| {
            output.replace(secret, "[REDACTED]")
        });
    if let Some(marker) = output.find("\n[OUTPUT TRUNCATED]") {
        let mut body = output[..marker].to_owned();
        for secret in secrets.iter().filter(|secret| !secret.is_empty()) {
            let fragment = secret
                .char_indices()
                .map(|(index, character)| index + character.len_utf8())
                .filter(|length| *length < secret.len())
                .filter_map(|length| secret.get(..length))
                .filter(|prefix| body.ends_with(prefix))
                .max_by_key(|prefix| prefix.len());
            if let Some(fragment) = fragment {
                body.truncate(body.len() - fragment.len());
                body.push_str("[REDACTED]");
            }
        }
        output = format!("{body}{}", &output[marker..]);
    }
    output
}

/// Require a successful process result.
///
/// # Errors
///
/// Returns `AINFRA-E500` with already-redacted diagnostic text.
pub fn require_success(result: ProcessResult) -> Result<ProcessResult, AinfraError> {
    if result.return_code == 0 {
        return Ok(result);
    }
    Err(AinfraError::dependency(format!(
        "underlying command failed ({}): {}",
        result.return_code,
        result.stderr.trim()
    )))
}

fn read_bounded<R: Read>(mut reader: R, limit: usize) -> std::io::Result<Vec<u8>> {
    let mut retained = Vec::new();
    let mut truncated = false;
    let mut buffer = [0_u8; 8192];
    loop {
        let count = reader.read(&mut buffer)?;
        if count == 0 {
            break;
        }
        let remaining = limit.saturating_sub(retained.len());
        retained.extend_from_slice(&buffer[..count.min(remaining)]);
        truncated |= count > remaining;
    }
    if truncated {
        retained.extend_from_slice(b"\n[OUTPUT TRUNCATED]");
    }
    Ok(retained)
}

fn join_output(
    handle: thread::JoinHandle<std::io::Result<Vec<u8>>>,
    stream: &str,
) -> Result<Vec<u8>, AinfraError> {
    handle
        .join()
        .map_err(|_| AinfraError::dependency(format!("{stream} reader panicked")))?
        .map_err(|error| AinfraError::dependency(format!("cannot read {stream}: {error}")))
}

/// Ensure a working directory exists before process execution.
///
/// # Errors
///
/// Returns `AINFRA-E500` when the path is not a directory.
pub fn require_working_directory(path: &Path) -> Result<(), AinfraError> {
    if path.is_dir() {
        Ok(())
    } else {
        Err(AinfraError::dependency(format!(
            "working directory is missing: {}",
            path.display()
        )))
    }
}
