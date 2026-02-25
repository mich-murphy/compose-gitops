# =============================================================================
# compose-gitops container image is built with Nix (NixOS/nixpkgs) for a
# minimal, reproducible, layered image—no Alpine or glibc base.
# =============================================================================
#
# Option A — Build with Nix directly (recommended):
#
#   cd compose-gitops
#   nix build .#docker
#   docker load < result
#
# Option B — Build using this Dockerfile (produces a builder; then run it):
#
#   docker build -t compose-gitops-builder .
#   docker run --rm -v "$(pwd)/out:/out" compose-gitops-builder \
#     sh -c "nix build .#docker -o /tmp/r && cp -L /tmp/r /out/image.tar.gz"
#   docker load < out/image.tar.gz
#
# First-time: if Nix reports a vendor hash mismatch, set vendorHash in flake.nix
# to the suggested "got: sha256-..." value.
# =============================================================================

FROM nixos/nix:latest
WORKDIR /app
COPY . .
# Ensure the flake can be read (no git required for build)
RUN nix flake show . 2>/dev/null || true
# Default: print usage
CMD ["sh", "-c", "echo 'Run with: docker run --rm -v $(pwd)/out:/out compose-gitops-builder sh -c \"nix build .#docker -o /tmp/r && cp -L /tmp/r /out/image.tar.gz\"' && echo 'Then: docker load < out/image.tar.gz'"]
