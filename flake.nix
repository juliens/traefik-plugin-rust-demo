{
  description = "Demo plugin with rustt";

  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";
    fenix = {
      url = "https://flakehub.com/f/nix-community/fenix/0.1.*.tar.gz";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    naersk = {
      url = "https://flakehub.com/f/nix-community/naersk/0.1.*.tar.gz";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    flake-compat.url = "https://flakehub.com/f/edolstra/flake-compat/*.tar.gz";
    flake-schemas.url = "https://flakehub.com/f/DeterminateSystems/flake-schemas/*.tar.gz";
  };

  outputs = { self, ... }@inputs:
    let
      pkgName = (self.lib.fromToml ./Cargo.toml).package.name;
      supportedSystems = [ "aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux" ];
      forAllSystems = f: inputs.nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import inputs.nixpkgs { inherit system; overlays = [ self.overlays.default ]; };
        inherit system;
      });
      rustWasmTarget = "wasm32-wasip1";
    in
    {
      overlays.default = final: prev: rec {
        system = final.stdenv.hostPlatform.system;

        # Builds a Rust toolchain from rust-toolchain.toml
        rustToolchain = with inputs.fenix.packages.${system};
          combine [
            latest.rustc
            latest.cargo
            targets.${rustWasmTarget}.latest.rust-std
          ];

        buildRustWasiWasm = self.lib.buildRustWasiWasm final;
        buildTraefikPlugin = self.lib.buildTraefikPlugin final;
      };

      # Development environments
      devShells = forAllSystems ({ pkgs, system }: {
        default =
          let
            helpers = with pkgs; [ direnv jq ];
          in
          pkgs.mkShell {
            packages = helpers ++ (with pkgs; [
              rustToolchain # cargo, etc.
              cargo-edit # cargo add, cargo rm, etc.
            ]);
          };
      });

      packages = forAllSystems ({ pkgs, system }: rec {
        "${pkgName}" = pkgs.buildRustWasiWasm {
          name = pkgName;
          src = self;
        };

        default = self.packages.${system}.${pkgName};

        "${pkgName}-plugin" = pkgs.stdenv.mkDerivation {
              name = "${pkgName}-plugin";
              src = ./.;

              buildInputs = [ pkgs.zip ];
               buildPhase = ''
                            mkdir -p $out/{lib,bin}
                            cp ${default}/lib/http-wasm-header.wasm $out/lib/http-wasm-header.wasm
                            cp $src/.traefik.yaml $out/lib/.traefik.yaml
                            zip -j $out/bin/plugin.zip $out/lib/http-wasm-header.wasm $out/lib/.traefik.yaml
                         '';
          };
      });

      lib = {
        # Helper function for reading TOML files
        fromToml = file: builtins.fromTOML (builtins.readFile file);

        handleArgs =
          { name ? null
          , src ? self
          , cargoToml ? ./Cargo.toml
          }:
          let
            meta = (self.lib.fromToml ./Cargo.toml).package;
            pkgName = if name == null then meta.name else name;
            pkgSrc = builtins.path { path = src; name = "${pkgName}-source"; };
          in
          {
            inherit (meta) name;
            inherit pkgName;
            src = pkgSrc;
            inherit cargoToml;
          };

        buildRustWasiWasm = pkgs: { name, src }:
          let
            naerskLib = pkgs.callPackage inputs.naersk {
              cargo = pkgs.rustToolchain;
              rustc = pkgs.rustToolchain;
            };
          in
          naerskLib.buildPackage {
            inherit name src;
            CARGO_BUILD_TARGET = rustWasmTarget;
            buildInputs = with pkgs; [ wabt ];
            postInstall = ''
              mkdir -p $out/lib
              wasm-strip $out/bin/${name}.wasm -o $out/lib/${name}.wasm
              rm -rf $out/bin
              wasm-validate $out/lib/${name}.wasm
            '';
          };
      };
    };
}