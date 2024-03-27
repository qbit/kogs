{
  description = "kogs: koreader sync server";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs =
    { self
    , nixpkgs
    ,
    }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      overlay = _: prev: { inherit (self.packages.${prev.system}) kogs; };
      nixosModule = import ./module.nix;
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          kogs = pkgs.buildGoModule {
            pname = "kogs";
            version = "v0.1.0";
            src = ./.;

            vendorHash = "sha256-8AviacBPdpvuII/2symR1IgcT0Bf5OL6Do/6Go8TD1A=";
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.kogs);
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              nix run github:qbit/xin#flake-warn
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [ git go gopls go-tools ];
          };
        });
    };
}
