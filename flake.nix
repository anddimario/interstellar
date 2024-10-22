{
  inputs = {
    # nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    { systems, nixpkgs, ... }@inputs:
    let
      eachSystem = f: nixpkgs.lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
    in
    {
      devShells = eachSystem (pkgs: {
        default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.air
            # pkgs.hurl
            # pkgs.firecracker
            # pkgs.firectl
            # pkgs.podman
            pkgs.gh
            pkgs.siege
          ];
          
          shellHook = ''
            alias gcam="git commit -am"
            alias gp="git push"
            alias gst="git status"
            alias docker="podman"
            export GOPATH=$HOME/go
            export PATH=$GOPATH/bin:$PATH
            echo "GOPATH is set to $GOPATH"
            zsh
            #code .
          '';

        };
      });
    };
}
