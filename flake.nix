{
  description = "gossip-glomers";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = github:numtide/flake-utils;
    maelstrom-nix.url = "github:stefanobaghino/maelstrom-nix";
  };

  outputs = { self, nixpkgs, flake-utils, maelstrom-nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        maelstrom = maelstrom-nix.packages.${system}.default;
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.nixpkgs-fmt
            pkgs.go
            pkgs.gotools
            pkgs.gopls
            pkgs.godef
            pkgs.go-outline
            maelstrom
          ];
        };
      }
    );
}
