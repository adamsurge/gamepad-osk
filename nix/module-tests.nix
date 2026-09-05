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
  overridePackage = pkgs.hello;

  nixosConfig =
    (nixpkgs.lib.nixosSystem {
      system = pkgs.stdenv.hostPlatform.system;
      modules = [
        nixosModule
        {
          programs.gamepad-osk = {
            enable = true;
            package = overridePackage;
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
          package = overridePackage;
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

  defaultNixosConfig =
    (nixpkgs.lib.nixosSystem {
      system = pkgs.stdenv.hostPlatform.system;
      modules = [
        nixosModule
        {
          programs.gamepad-osk.enable = true;
          system.stateVersion = "26.05";
        }
      ];
    }).config;

  standaloneNixosConfig =
    (nixpkgs.lib.nixosSystem {
      system = pkgs.stdenv.hostPlatform.system;
      modules = [
        ./nixos-module.nix
        {
          programs.gamepad-osk.enable = true;
          system.stateVersion = "26.05";
        }
      ];
    }).config;

  disabledNixosConfig =
    (nixpkgs.lib.nixosSystem {
      system = pkgs.stdenv.hostPlatform.system;
      modules = [
        nixosModule
        { system.stateVersion = "26.05"; }
      ];
    }).config;

  defaultHomeConfig = home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = [
      homeManagerModule
      {
        home = {
          username = "tester";
          homeDirectory = "/home/tester";
          stateVersion = "26.05";
        };
        programs.gamepad-osk.enable = true;
      }
    ];
  };

  standaloneHomeConfig = home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = [
      ./home-manager-module.nix
      {
        home = {
          username = "tester";
          homeDirectory = "/home/tester";
          stateVersion = "26.05";
        };
        programs.gamepad-osk.enable = true;
      }
    ];
  };

  serviceHomeConfig = home-manager.lib.homeManagerConfiguration {
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
          service.enable = true;
        };
      }
    ];
  };

  disabledHomeConfig = home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = [
      homeManagerModule
      {
        home = {
          username = "tester";
          homeDirectory = "/home/tester";
          stateVersion = "26.05";
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
  generatedUnit = serviceHomeConfig.config.xdg.configFile."systemd/user/gamepad-osk.service".source;
  systemdBasicTarget = pkgs.writeText "basic.target" ''
    [Unit]
    Description=Basic User System
  '';
  systemdUserUnitPath = pkgs.runCommand "gamepad-osk-systemd-user-units" { } ''
    mkdir -p "$out"
    ln -s ${systemdBasicTarget} "$out/basic.target"
  '';
in
{
  nixos-module =
    assert lib.elem overridePackage nixosConfig.environment.systemPackages;
    assert lib.elem overridePackage nixosConfig.services.udev.packages;
    assert nixosConfig.programs.gamepad-osk.package == overridePackage;
    assert lib.elem "uinput" nixosConfig.boot.kernelModules;
    assert lib.elem package defaultNixosConfig.environment.systemPackages;
    assert defaultNixosConfig.programs.gamepad-osk.package == package;
    assert standaloneNixosConfig.programs.gamepad-osk.package.pname == "gamepad-osk";
    assert !(lib.elem package disabledNixosConfig.environment.systemPackages);
    assert !(lib.elem package disabledNixosConfig.services.udev.packages);
    assert !(lib.elem "uinput" disabledNixosConfig.boot.kernelModules);
    pkgs.runCommand "gamepad-osk-nixos-module-check" { } ''
      touch "$out"
    '';

  home-manager-module =
    assert lib.elem overridePackage homeConfig.config.home.packages;
    assert homeConfig.config.programs.gamepad-osk.package == overridePackage;
    assert lib.elem package defaultHomeConfig.config.home.packages;
    assert defaultHomeConfig.config.programs.gamepad-osk.package == package;
    assert standaloneHomeConfig.config.programs.gamepad-osk.package.pname == "gamepad-osk";
    assert !(lib.elem package disabledHomeConfig.config.home.packages);
    assert !(disabledHomeConfig.config.xdg.configFile ? "gamepad-osk/config");
    assert !(disabledHomeConfig.config.systemd.user.services ? gamepad-osk);
    assert unit.After == [ "graphical-session.target" ];
    assert unit.PartOf == [ "graphical-session.target" ];
    assert install.WantedBy == [ "graphical-session.target" ];
    assert customTargetUnit.After == [ "niri.service" ];
    assert customTargetUnit.PartOf == [ "niri.service" ];
    assert customTargetInstall.WantedBy == [ "niri.service" ];
    assert service.ExecStart == [ "${overridePackage}/bin/gamepad-osk --daemon" ];
    assert service.ExecStartPre == "${pkgs.coreutils}/bin/sleep 3";
    pkgs.runCommand "gamepad-osk-home-manager-module-check" { nativeBuildInputs = [ pkgs.systemd ]; } ''
      grep -F '[theme]' ${generatedConfig}
      grep -F 'name=matrix' ${generatedConfig}
      grep -F '[gamepad.buttons]' ${generatedConfig}
      grep -F 'press=a' ${generatedConfig}
      export XDG_RUNTIME_DIR="$TMPDIR/runtime"
      mkdir -p "$XDG_RUNTIME_DIR"
      export SYSTEMD_UNIT_PATH="${systemdUserUnitPath}:${pkgs.systemd}/lib/systemd/user"
      systemd-analyze --user verify \
        ${package}/share/systemd/user/gamepad-osk.service \
        ${generatedUnit}
      touch "$out"
    '';
}
