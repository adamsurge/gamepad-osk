{
  description = "Gamepad-controlled on-screen keyboard for Linux";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      home-manager,
      ...
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          promptfont = pkgs.callPackage ./nix/promptfont.nix { };
          gamepad-osk = pkgs.callPackage ./nix/package.nix { inherit promptfont; };
        in
        {
          inherit gamepad-osk promptfont;
          default = gamepad-osk;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/gamepad-osk";
          meta.description = "Gamepad-controlled on-screen keyboard for Linux";
        };
      });

      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          promptfont = self.packages.${system}.promptfont;
          package = self.packages.${system}.gamepad-osk;
          moduleChecks = import ./nix/module-tests.nix {
            inherit pkgs nixpkgs home-manager;
            inherit package;
            nixosModule = self.nixosModules.default;
            homeManagerModule = self.homeManagerModules.default;
          };
        in
        {
          package-artifacts = pkgs.callPackage ./nix/package-tests.nix {
            inherit package promptfont;
          };
          nix-lint =
            pkgs.runCommand "gamepad-osk-nix-lint"
              {
                nativeBuildInputs = [
                  pkgs.deadnix
                  pkgs.statix
                ];
              }
              ''
                deadnix --fail ${./flake.nix} ${./nix}
                statix check ${./flake.nix}
                statix check ${./nix}
                touch "$out"
              '';
          inherit (moduleChecks) nixos-module home-manager-module;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          package = self.packages.${system}.default;
        in
        {
          default = pkgs.mkShell {
            inputsFrom = [ package ];
            packages = [ pkgs.go_1_26 ];
          };
        }
      );

      formatter = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        pkgs.writeShellScriptBin "nixfmt" ''
          ${pkgs.nixfmt}/bin/nixfmt $(${pkgs.git}/bin/git ls-files -- '*.nix')
        ''
      );

      nixosModules = rec {
        gamepad-osk =
          { lib, pkgs, ... }:
          {
            imports = [ ./nix/nixos-module.nix ];
            programs.gamepad-osk.package =
              lib.mkDefault
                self.packages.${pkgs.stdenv.hostPlatform.system}.default;
          };
        default = gamepad-osk;
      };

      homeManagerModules = rec {
        gamepad-osk =
          { lib, pkgs, ... }:
          {
            imports = [ ./nix/home-manager-module.nix ];
            programs.gamepad-osk.package =
              lib.mkDefault
                self.packages.${pkgs.stdenv.hostPlatform.system}.default;
          };
        default = gamepad-osk;
      };
    };
}
