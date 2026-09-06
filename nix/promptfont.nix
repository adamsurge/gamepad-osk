{
  fetchzip,
  lib,
  stdenvNoCC,
}:

stdenvNoCC.mkDerivation (finalAttrs: {
  pname = "promptfont";
  version = "1.15";

  src = fetchzip {
    url = "https://codeberg.org/shinmera/promptfont/releases/download/v${finalAttrs.version}/promptfont.zip";
    hash = "sha256-k+SLlFQ/gUarsmBVBhT+IxMsKTytHhU3PVqwbLOaCKU=";
    stripRoot = false;
  };

  dontBuild = true;

  installPhase = ''
    runHook preInstall

    install -Dm444 promptfont.ttf "$out/share/fonts/truetype/promptfont/promptfont.ttf"
    install -Dm444 LICENSE.txt "$out/share/licenses/${finalAttrs.pname}/LICENSE.txt"
    install -Dm444 README.md "$out/share/doc/${finalAttrs.pname}/README.md"

    runHook postInstall
  '';

  meta = {
    description = "Controller-agnostic icon font";
    homepage = "https://codeberg.org/shinmera/promptfont";
    license = lib.licenses.ofl;
    platforms = lib.platforms.all;
  };
})
