{
  pkgs,
  nixpkgs,
  home-manager,
  package,
  nixosModule,
  homeManagerModule,
}:

let
  lib = pkgs.lib;

  nixosConfig = (nixpkgs.lib.nixosSystem {
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

  generatedConfig = homeConfig.config.xdg.configFile."gamepad-osk/config".source;
  service = homeConfig.config.systemd.user.services.gamepad-osk.Service;
in {
  nixos-module =
    assert lib.elem package nixosConfig.environment.systemPackages;
    assert lib.elem package nixosConfig.services.udev.packages;
    pkgs.runCommand "gamepad-osk-nixos-module-check" { } ''
      touch "$out"
    '';

  home-manager-module =
    assert lib.elem package homeConfig.config.home.packages;
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
