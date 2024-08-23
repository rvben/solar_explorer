{
  description = "A flake to install the Solar Exporter application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs }: {
    packages =
      let
        commonDerivation = { pkgs, system }: pkgs.stdenv.mkDerivation {
          pname = "solar_exporter";
          version = "1.0.1";

          src = ./.;

          buildInputs = [ pkgs.go_1_23 ];

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
      in
      {
        x86_64-linux = commonDerivation { pkgs = import nixpkgs { system = "x86_64-linux"; }; system = "x86_64-linux"; };
        aarch64-darwin = commonDerivation { pkgs = import nixpkgs { system = "aarch64-darwin"; }; system = "aarch64-darwin"; };
      };

    defaultPackage.x86_64-linux = self.packages.x86_64-linux;
    defaultPackage.aarch64-darwin = self.packages.aarch64-darwin;

    nixosModules = {
      solar_exporter = { pkgs, ... }: {
        nixpkgs.overlays = [
          (final: prev: {
            solar_exporter = self.packages.${prev.system};
          })
        ];

        imports = [ ./nixosModules/solar_exporter.nix ];
      };
    };
  };
}
