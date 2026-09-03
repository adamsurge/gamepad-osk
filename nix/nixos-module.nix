{ config, lib, pkgs, ... }:

let
  cfg = config.programs.gamepad-osk;
in {
  options.programs.gamepad-osk = {
    enable = lib.mkEnableOption "gamepad-osk";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.callPackage ./package.nix { };
      defaultText = lib.literalExpression "pkgs.callPackage ./nix/package.nix { }";
      description = "gamepad-osk package to install.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];
    services.udev.packages = [ cfg.package ];
  };
}
