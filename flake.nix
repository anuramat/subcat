{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };
  outputs =
    inputs:
    let
      name = "subcat";
      system = "x86_64-linux"; # TODO all systems
      pkgs = (import inputs.nixpkgs) { inherit system; };
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        pname = name;
        version = "0.0.1";
        src = ./.;
        vendorHash = "sha256-KxT0ZHRZURa9DWjvVSAt3FOZq8HGy1ocwQy1Y0E+iTw=";
        meta = {
          mainProgram = name;
        };
      };
    };
}
