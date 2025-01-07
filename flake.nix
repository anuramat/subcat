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
        vendorHash = "sha256-Koc5Z//NrrPyFVR+XchoL2DIQc9/vUCr9XUDPucIFOA=";
        meta = {
          mainProgram = name;
        };
      };
    };
}
