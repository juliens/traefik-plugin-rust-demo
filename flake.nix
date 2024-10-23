{
  description = "Rust development environment";

  # Flake inputs
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    rust-overlay.url = "https://github.com/oxalica/rust-overlay/archive/master.tar.gz";
  };

  # Flake outputs
  outputs = { self, nixpkgs, rust-overlay}:
  let
      # Systems supported
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      # Helper to provide system-specific attributes
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        pkgs = import nixpkgs { inherit system;
        overlays = [ (import rust-overlay) ];
      };
      });



  in
  {
      # Development environment output
      devShells = forAllSystems ({ pkgs }: 
      let
        rustNightly = pkgs.rust-bin.nightly."2024-10-22".default.override {
          targets = [ "wasm32-wasip1" ];
          extensions = [ "rust-src" "rust-std" "cargo" "rustc" ];
        };
      in
      {
        default = pkgs.mkShell {
          shellHook = ''
          '';

  # Pour la compilation C
  #CFLAGS = "-I${pkgs.glibc.dev}/include";
  #LDFLAGS = "-L${pkgs.glibc}/lib";

  # Pour rust-bindgen
  #LIBCLANG_PATH = "${pkgs.llvmPackages.libclang.lib}/lib";

  # Debug
  #RUST_BACKTRACE = 1;
  packages = (with pkgs; [
    rustNightly
    #wasm-pack
    #wasmtime


    # Dépendances de build essentielles
#   clang
#   llvm
#   gcc
#   binutils

#   # Outils de développement
#   pkg-config
#   openssl.dev

#   # Bibliothèques système nécessaires
#   glibc.dev

#   # Pour la cross-compilation
#   lld_13
#   cmake
  ]) ++ pkgs.lib.optionals pkgs.stdenv.isDarwin (with pkgs; [ libiconv ]);
};
      });
    };
  }
