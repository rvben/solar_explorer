{
  description = "A flake to install the Go application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs }: {
    packages = {
      x86_64-linux = let
        pkgs = import nixpkgs { system = "x86_64-linux"; };
      in pkgs.stdenv.mkDerivation {
        pname = "solar_exporter";
        version = "1.0.0";

        src = ./.;

        buildInputs = [ pkgs.go ];

        buildPhase = ''
          export GOPATH=$TMPDIR/go-path
          export GOCACHE=$TMPDIR/go-cache
          export GOMODCACHE=$TMPDIR/go-mod-cache
          mkdir -p $GOPATH $GOCACHE $GOMODCACHE
          go build -o solar_exporter .
        '';

        installPhase = ''
          mkdir -p $out/bin
          cp solar_exporter $out/bin/
        '';

        meta = with pkgs.lib; {
          description = "A Go application for solar metrics export";
          license = licenses.mit;
          maintainers = with maintainers; [ rvben ];
        };
      };

      aarch64-darwin = let
        pkgs = import nixpkgs { system = "aarch64-darwin"; };
      in pkgs.stdenv.mkDerivation {
        pname = "solar_exporter";
        version = "1.0.0";

        src = ./.;

        buildInputs = [ pkgs.go ];

        buildPhase = ''
          export GOPATH=$TMPDIR/go-path
          export GOCACHE=$TMPDIR/go-cache
          export GOMODCACHE=$TMPDIR/go-mod-cache
          mkdir -p $GOPATH $GOCACHE $GOMODCACHE
          go build -o solar_exporter .
        '';

        installPhase = ''
          mkdir -p $out/bin
          cp solar_exporter $out/bin/
        '';

        meta = with pkgs.lib; {
          description = "A Go application for solar metrics export";
          license = licenses.mit;
          maintainers = with maintainers; [ rvben ];
        };
      };
    };

    defaultPackage.x86_64-linux = self.packages.x86_64-linux;
    defaultPackage.aarch64-darwin = self.packages.aarch64-darwin;
  };
}