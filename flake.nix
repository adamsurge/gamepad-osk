{
  description = "Gamepad-controlled on-screen keyboard for Linux";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs@{ self, nixpkgs, home-manager, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          gamepad-osk = pkgs.callPackage ./nix/package.nix { };
        in {
          inherit gamepad-osk;
          default = gamepad-osk;
        });

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/gamepad-osk";
          meta.description = "Gamepad-controlled on-screen keyboard for Linux";
        };
      });

      checks = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          moduleChecks = import ./nix/module-tests.nix {
            inherit pkgs nixpkgs home-manager;
            package = self.packages.${system}.default;
            nixosModule = self.nixosModules.default;
            homeManagerModule = self.homeManagerModules.default;
          };
        in {
          gamepad-osk = self.packages.${system}.default;
          inherit (moduleChecks) nixos-module home-manager-module;
        });

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          package = self.packages.${system}.default;
        in {
          default = pkgs.mkShell {
            inputsFrom = [ package ];
            packages = [ pkgs.go_1_26 ];
            GAMEPAD_OSK_FONT_DIRS = "${pkgs.dejavu_fonts}/share/fonts/truetype/DejaVu";
          };
        });

      nixosModules.default = import ./nix/nixos-module.nix;
      homeManagerModules.default = import ./nix/home-manager-module.nix;
    };
}
