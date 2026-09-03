{ config, lib, pkgs, ... }:

let
  cfg = config.programs.gamepad-osk;
  iniFormat = pkgs.formats.ini { };
in {
  options.programs.gamepad-osk = {
    enable = lib.mkEnableOption "gamepad-osk";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.callPackage ./package.nix { };
      defaultText = lib.literalExpression "pkgs.callPackage ./nix/package.nix { }";
      description = "gamepad-osk package to use.";
    };

    settings = lib.mkOption {
      inherit (iniFormat) type;
      default = { };
      example = {
        theme.name = "matrix";
        window.position = "bottom";
        gamepad.toggle_combo = "guide+a";
        "gamepad.buttons".press = "a";
      };
      description = "Settings written to gamepad-osk's INI configuration.";
    };

    service.enable = lib.mkEnableOption "gamepad-osk graphical-session user service";
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];

    xdg.configFile."gamepad-osk/config".source =
      iniFormat.generate "gamepad-osk-config" cfg.settings;

    systemd.user.services.gamepad-osk = lib.mkIf cfg.service.enable {
      Unit = {
        Description = "gamepad-osk - Gamepad on-screen keyboard";
        After = [ "graphical-session.target" ];
      };
      Service = {
        Type = "simple";
        ExecStartPre = "${pkgs.coreutils}/bin/sleep 3";
        ExecStart = "${cfg.package}/bin/gamepad-osk --daemon";
        Restart = "on-failure";
        RestartSec = 5;
      };
      Install.WantedBy = [ "graphical-session.target" ];
    };
  };
}
