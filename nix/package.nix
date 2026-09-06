{
  buildGo126Module,
  coreutils,
  dejavu_fonts,
  fontconfig,
  lib,
  libX11,
  pkg-config,
  promptfont,
  sdl3,
  sdl3-ttf,
  wayland,
}:

buildGo126Module (finalAttrs: {
  pname = "gamepad-osk";
  version = "2.1.1";

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../config.example
      ../gamepad-osk.service
      ../gamepad-osk.udev
      ../go.mod
      ../go.sum
      (lib.fileset.fileFilter (file: file.hasExt "c") ../.)
      (lib.fileset.fileFilter (file: file.hasExt "go") ../.)
      (lib.fileset.fileFilter (file: file.hasExt "h") ../.)
    ];
  };

  vendorHash = null;
  strictDeps = true;

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [
    fontconfig
    sdl3
    sdl3-ttf
    wayland
    libX11
  ];

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${finalAttrs.version}"
  ];
  doCheck = true;

  postInstall = ''
    install -Dm644 config.example "$out/share/gamepad-osk/config"
    install -Dm644 gamepad-osk.udev "$out/lib/udev/rules.d/80-gamepad-osk.rules"
    install -Dm644 gamepad-osk.service "$out/share/systemd/user/gamepad-osk.service"
    substituteInPlace "$out/share/systemd/user/gamepad-osk.service" \
      --replace-fail /bin/sleep ${coreutils}/bin/sleep \
      --replace-fail /usr/bin/gamepad-osk "$out/bin/gamepad-osk"
    mkdir -p "$out/share/gamepad-osk/fonts"
    ln -s ${promptfont}/share/fonts/truetype/promptfont/promptfont.ttf \
      "$out/share/gamepad-osk/fonts/promptfont.ttf"
    ln -s ${dejavu_fonts}/share/fonts/truetype/DejaVuSans.ttf \
      "$out/share/gamepad-osk/fonts/DejaVuSans.ttf"
  '';

  meta = {
    description = "Gamepad-controlled on-screen keyboard for Linux";
    homepage = "https://github.com/0x90shell/gamepad-osk";
    license = lib.licenses.mit;
    mainProgram = "gamepad-osk";
    platforms = lib.platforms.linux;
  };
})
