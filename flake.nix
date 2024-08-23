{
  description = "A flake to install the Go application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs }: {
    packages = let
      pkgs = import nixpkgs { system = "x86_64-linux"; };
    in {
      default = pkgs.stdenv.mkDerivation {
        pname = "solar_exporter";
        version = "1.0.0";

        src = ./.;

        buildInputs = [ pkgs.go ];

        buildPhase = ''
          mkdir -p $out/bin
          go build -o $out/bin/solar_exporter .
        '';

        installPhase = ''
          mkdir -p $out/bin
          cp $out/bin/solar_exporter $out/bin/
        '';

        meta = with pkgs.lib; {
          description = "A Go application for solar metrics export";
          license = licenses.mit;
          maintainers = with maintainers; [ rvben ];
        };
      };
    };
  };
}