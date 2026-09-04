{
  pkgs,
  nixpkgs,
  home-manager,
  package,
  nixosModule,
  homeManagerModule,
}:

let
  inherit (pkgs) lib;

  nixosConfig =
    (nixpkgs.lib.nixosSystem {
      system = pkgs.stdenv.hostPlatform.system;
      modules = [
        nixosModule
        {
          programs.gamepad-osk = {
            enable = true;
            inherit package;
          };
          system.stateVersion = "26.05";
        }
      ];
    }).config;

  homeConfig = home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = [
      homeManagerModule
      {
        home = {
          username = "tester";
          homeDirectory = "/home/tester";
          stateVersion = "26.05";
        };
        programs.gamepad-osk = {
          enable = true;
          inherit package;
          settings = {
            theme.name = "matrix";
            window.position = "bottom";
            gamepad.toggle_combo = "guide+a";
            "gamepad.buttons".press = "a";
          };
          service.enable = true;
        };
      }
    ];
  };

  customTargetHomeConfig = home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = [
      homeManagerModule
      {
        home = {
          username = "tester";
          homeDirectory = "/home/tester";
          stateVersion = "26.05";
        };
        programs.gamepad-osk = {
          enable = true;
          inherit package;
          service = {
            enable = true;
            systemdTargets = [ "niri.service" ];
          };
        };
      }
    ];
  };

  generatedConfig = homeConfig.config.xdg.configFile."gamepad-osk/config".source;
  unit = homeConfig.config.systemd.user.services.gamepad-osk.Unit;
  service = homeConfig.config.systemd.user.services.gamepad-osk.Service;
  install = homeConfig.config.systemd.user.services.gamepad-osk.Install;
  customTargetUnit = customTargetHomeConfig.config.systemd.user.services.gamepad-osk.Unit;
  customTargetInstall = customTargetHomeConfig.config.systemd.user.services.gamepad-osk.Install;
in
{
  nixos-module =
    assert lib.elem package nixosConfig.environment.systemPackages;
    assert lib.elem package nixosConfig.services.udev.packages;
    pkgs.runCommand "gamepad-osk-nixos-module-check" { } ''
      touch "$out"
    '';

  home-manager-module =
    assert lib.elem package homeConfig.config.home.packages;
    assert unit.After == [ "graphical-session.target" ];
    assert unit.PartOf == [ "graphical-session.target" ];
    assert install.WantedBy == [ "graphical-session.target" ];
    assert customTargetUnit.After == [ "niri.service" ];
    assert customTargetUnit.PartOf == [ "niri.service" ];
    assert customTargetInstall.WantedBy == [ "niri.service" ];
    assert service.ExecStart == [ "${package}/bin/gamepad-osk --daemon" ];
    assert service.ExecStartPre == "${pkgs.coreutils}/bin/sleep 3";
    pkgs.runCommand "gamepad-osk-home-manager-module-check" { } ''
      grep -F '[theme]' ${generatedConfig}
      grep -F 'name=matrix' ${generatedConfig}
      grep -F '[gamepad.buttons]' ${generatedConfig}
      grep -F 'press=a' ${generatedConfig}
      touch "$out"
    '';
}
