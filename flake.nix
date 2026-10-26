# Development environment and container image build for compose-gitops.
# - nix develop / direnv: dev shell (Go, Docker, etc.)
# - nix build .#docker: container image
# - nix build .#default: CLI binary
{
  description = "compose-gitops development and image build";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    allSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
      "x86_64-darwin"
    ];
    # Systems we build the Docker image for (Linux only)
    linuxSystems = [ "x86_64-linux" "aarch64-linux" ];

    forAllSystems = f:
      nixpkgs.lib.genAttrs allSystems (system:
        f {
          pkgs = import nixpkgs { inherit system; };
        });

    forLinuxSystems = f:
      nixpkgs.lib.genAttrs linuxSystems (system:
        f {
          pkgs = import nixpkgs { inherit system; };
        });
  in {
    devShells = forAllSystems ({pkgs}: {
      default = pkgs.mkShell {
        packages = [
          pkgs.go_1_26
          pkgs.gopls
          pkgs.delve
          pkgs.docker
          pkgs.docker-compose
          pkgs.git
          pkgs.cacert
          pkgs.golangci-lint
        ];

        env = {
          CGO_ENABLED = "0";
          SSL_CERT_FILE = "${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt";
          NIX_SSL_CERT_FILE = "${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt";
        };

        shellHook = ''
          export PATH="$PATH:$(go env GOPATH)/bin"
        '';
      };
    });

    packages = forLinuxSystems ({pkgs}: let
      compose-gitops = pkgs.buildGoModule {
        pname = "compose-gitops";
        version = "0.1.0";
        src = ./.;
        subPackages = [ "cmd/compose-gitops" ];
        vendorHash = "sha256-SQEmlwrZi5kODb/LKzKQ9mKykDZ7nHIM3KsrJqOPkYM=";
        env.CGO_ENABLED = "0";
        ldflags = [ "-s" "-w" ];
      };
      imageContents = [
        compose-gitops
        pkgs.git
        pkgs.cacert
        pkgs.iana-etc
      ];
    in {
      default = compose-gitops;
      compose-gitops = compose-gitops;
      docker = pkgs.dockerTools.buildLayeredImage {
        name = "compose-gitops";
        tag = "latest";
        contents = imageContents;
        config = {
          Entrypoint = [ "/bin/compose-gitops" ];
          Env = [
            "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
            "NIX_SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
          ];
          ExposedPorts = { "8080/tcp" = {}; };
        };
      };
    });
  };
}
