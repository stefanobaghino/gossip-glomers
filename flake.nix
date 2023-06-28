{
  description = "gossip-glomers";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = github:numtide/flake-utils;
    maelstrom.url = "path:./nix/maelstrom";
  };

  outputs = { self, nixpkgs, flake-utils, maelstrom }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.nixpkgs-fmt
            pkgs.go
            pkgs.gotools
            maelstrom.packages.${system}.default
          ];
        };
      }
    );
}
