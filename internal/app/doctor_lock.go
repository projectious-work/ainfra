package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/projectious-work/ainfra/internal/doctor"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/source"
)

func deploymentTemplateFacts(
	deployment project.Deployment,
	cacheRoot string,
) []doctor.DeploymentTemplateFact {
	lockPath := filepath.Join(deployment.Target.Root, lockfile.Filename)
	document, err := lockfile.Read(lockPath)
	if err != nil {
		message := "template lock is invalid"
		if errors.Is(err, os.ErrNotExist) {
			message = "template lock is missing"
		}
		facts := []doctor.DeploymentTemplateFact{
			lockFact("template.lock", "AINFRA-E2310", lockPath, "fail", message,
				"Run 'ainfra template lock' for this deployment."),
		}
		for _, pending := range []struct{ id, code, message string }{
			{"template.source-binding", "AINFRA-E2311", "source binding unavailable until the lock is valid"},
			{"template.cache", "AINFRA-E2312", "verified cache entry unavailable until the lock is valid"},
			{"template.digest", "AINFRA-E2313", "cached digest unavailable until the lock is valid"},
			{"template.source-content", "AINFRA-E2314", "source content comparison unavailable until the lock is valid"},
		} {
			facts = append(facts, lockFact(pending.id, pending.code, lockPath, "skip", pending.message,
				"Repair the lock with 'ainfra template lock' or 'ainfra template update'."))
		}
		return facts
	}
	facts := []doctor.DeploymentTemplateFact{
		lockFact("template.lock", "AINFRA-E2310", lockPath, "pass",
			"template lock is structurally valid", ""),
	}
	requestedSource, parseErr := source.Parse(deployment.Template.Source, deployment.Template.Ref)
	if parseErr != nil || document.Template.Source != requestedSource.Display ||
		document.Template.RequestedRef != requestedSource.RequestedRef {
		facts = append(facts, lockFact(
			"template.source-binding", "AINFRA-E2311", lockPath, "fail",
			"deployment source or requested ref differs from the template lock",
			"Review the source change, then run 'ainfra template update'.",
		))
	} else {
		facts = append(facts, lockFact(
			"template.source-binding", "AINFRA-E2311", lockPath, "pass",
			"deployment source and requested ref match the template lock", "",
		))
	}
	cachePath := filepath.Join(
		cacheRoot, "templates", "sha256", strings.TrimPrefix(document.Template.Digest, "sha256:"),
	)
	observed, digestErr := source.TreeDigest(cachePath)
	if digestErr != nil {
		facts = append(facts,
			lockFact("template.cache", "AINFRA-E2312", cachePath, "fail",
				"locked template cache entry is missing or unsafe",
				"Rerun 'ainfra template lock' to restore verified cached content."),
			lockFact("template.digest", "AINFRA-E2313", cachePath, "skip",
				"cached digest unavailable because the cache entry is invalid",
				"Restore the cache with 'ainfra template lock'."),
		)
	} else {
		facts = append(facts, lockFact("template.cache", "AINFRA-E2312", cachePath, "pass",
			"locked template cache entry is present and safe", ""))
		status, message, next := "pass", "cached template digest matches the lock", ""
		if observed != document.Template.Digest {
			status, message = "fail", "cached template digest differs from the lock"
			next = "Remove the poisoned cache entry and rerun 'ainfra template lock'."
		}
		facts = append(facts, lockFact("template.digest", "AINFRA-E2313", cachePath, status, message, next))
	}
	facts = append(facts, sourceContentFact(deployment, requestedSource, document, lockPath))
	return facts
}

func sourceContentFact(
	deployment project.Deployment,
	reference source.Reference,
	document lockfile.Document,
	lockPath string,
) doctor.DeploymentTemplateFact {
	if reference.Kind == source.KindGit {
		return lockFact("template.source-content", "AINFRA-E2314", lockPath, "skip",
			"mutable Git ref is not reacquired during doctor; locked commit and cache are checked",
			"Run 'ainfra template update' to resolve the Git ref explicitly.")
	}
	resolved, err := source.ResolveLocal(reference, deployment.Target.Root)
	if err != nil {
		return lockFact("template.source-content", "AINFRA-E2314", lockPath, "fail",
			"local template source is unavailable or outside approved roots",
			"Restore the approved local source and rerun doctor.")
	}
	digest, err := source.TreeDigest(resolved.Path)
	if err != nil {
		return lockFact("template.source-content", "AINFRA-E2314", resolved.Path, "fail",
			"local template source cannot be safely digested",
			"Remove unsafe content and run 'ainfra template update'.")
	}
	if digest != document.Template.Digest {
		return lockFact("template.source-content", "AINFRA-E2314", resolved.Path, "fail",
			"local template source content differs from the lock",
			"Review the content change, then run 'ainfra template update'.")
	}
	return lockFact("template.source-content", "AINFRA-E2314", resolved.Path, "pass",
		"local template source content matches the lock", "")
}

func lockFact(id, code, path, status, message, next string) doctor.DeploymentTemplateFact {
	return doctor.DeploymentTemplateFact{
		ID: id, Code: code, Path: path, Status: status, Message: message, NextAction: next,
	}
}
