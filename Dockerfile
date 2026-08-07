# Build the selected Linux binary first with scripts/build-targets.sh dist.
FROM scratch

ARG TARGETARCH
COPY --chown=65532:65532 dist/linux/${TARGETARCH}/ainfra /usr/local/bin/ainfra

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/ainfra"]
