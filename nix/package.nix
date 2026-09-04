{
  lib,
  buildGo126Module,
  coreutils,
  dejavu_fonts,
  libX11,
  makeWrapper,
  pkg-config,
  sdl3,
  sdl3-ttf,
  wayland,
}:

buildGo126Module rec {
  pname = "gamepad-osk";
  version = "2.1.1";

  src = lib.cleanSourceWith {
    src = ../.;
    filter =
      path: _type:
      let
        name = baseNameOf path;
      in
      name != ".git" && name != ".opencode" && name != "result";
  };

  vendorHash = null;

  nativeBuildInputs = [
    makeWrapper
    pkg-config
  ];
  buildInputs = [
    sdl3
    sdl3-ttf
    wayland
    libX11
  ];

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];
  doCheck = true;

  postInstall = ''
    install -Dm644 config.example "$out/share/gamepad-osk/config"
    install -Dm644 gamepad-osk.udev "$out/lib/udev/rules.d/80-gamepad-osk.rules"
    install -Dm644 gamepad-osk.service "$out/lib/systemd/user/gamepad-osk.service"
    substituteInPlace "$out/lib/systemd/user/gamepad-osk.service" \
      --replace-fail /bin/sleep ${coreutils}/bin/sleep \
      --replace-fail /usr/bin/gamepad-osk "$out/bin/gamepad-osk"
    wrapProgram "$out/bin/gamepad-osk" \
      --suffix GAMEPAD_OSK_FONT_DIRS : "${dejavu_fonts}/share/fonts/truetype/DejaVu"
  '';

  meta = {
    description = "Gamepad-controlled on-screen keyboard for Linux";
    homepage = "https://github.com/0x90shell/gamepad-osk";
    license = lib.licenses.mit;
    mainProgram = "gamepad-osk";
    platforms = lib.platforms.linux;
  };
}
