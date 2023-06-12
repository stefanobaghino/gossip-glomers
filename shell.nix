{ pkgs ? import <nixpkgs> { } }:

let
  maelstrom =
    pkgs.stdenv.mkDerivation {
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
      propagatedBuildInputs = [ pkgs.graphviz pkgs.gnuplot ];
    };
in
pkgs.mkShell {
  buildInputs = [
    pkgs.jdk17
    pkgs.go
    pkgs.gopls
    maelstrom
  ];
}
