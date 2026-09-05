{
  coreutils,
  dejavu_fonts,
  gnugrep,
  lib,
  package,
  promptfont,
  runCommand,
}:

assert package.meta.license == lib.licenses.mit;
assert package.meta.mainProgram == "gamepad-osk";
assert promptfont.meta.license == lib.licenses.ofl;
runCommand "gamepad-osk-package-artifacts"
  {
    nativeBuildInputs = [
      coreutils
      gnugrep
    ];
  }
  ''
    test -x ${package}/bin/gamepad-osk
    test -f ${package}/share/gamepad-osk/config
    test -f ${package}/lib/udev/rules.d/80-gamepad-osk.rules
    test -f ${package}/share/systemd/user/gamepad-osk.service
    test -f ${promptfont}/share/fonts/truetype/promptfont/promptfont.ttf
    test -f ${promptfont}/share/licenses/promptfont/LICENSE.txt
    test -f ${promptfont}/share/doc/promptfont/README.md

    test -L ${package}/share/gamepad-osk/fonts/promptfont.ttf
    test -L ${package}/share/gamepad-osk/fonts/DejaVuSans.ttf
    test "$(readlink -f ${package}/share/gamepad-osk/fonts/promptfont.ttf)" = \
      ${promptfont}/share/fonts/truetype/promptfont/promptfont.ttf
    test "$(readlink ${package}/share/gamepad-osk/fonts/DejaVuSans.ttf)" = \
      ${dejavu_fonts}/share/fonts/truetype/DejaVuSans.ttf

    grep -F ${coreutils}/bin/sleep ${package}/share/systemd/user/gamepad-osk.service
    grep -F ${package}/bin/gamepad-osk ${package}/share/systemd/user/gamepad-osk.service

    export HOME="$TMPDIR/home"
    export XDG_CONFIG_HOME="$HOME/.config"
    mkdir -p "$HOME"
    ${package}/bin/gamepad-osk --help >/dev/null
    cmp ${package}/share/gamepad-osk/config "$XDG_CONFIG_HOME/gamepad-osk/config"

    touch "$out"
  ''
