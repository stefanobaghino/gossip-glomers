{
  description = "maelstrom";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = github:numtide/flake-utils;
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.stdenv.mkDerivation {
          name = "maelstrom";
          version = "0.2.3";
          src = pkgs.fetchzip {
            url = "https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2";
            hash = "sha256-mE/FIHDLYd1lxAvECZGelZtbo0xkQgMroXro+xb9bMI=";
          };
          installPhase = ''
            mkdir -p $out/bin/lib
            cp maelstrom $out/bin
            cp lib/maelstrom.jar $out/bin/lib/
          '';
          propagatedBuildInputs = [ pkgs.graphviz pkgs.gnuplot pkgs.openjdk17 ];
        };
      }
    );
}
