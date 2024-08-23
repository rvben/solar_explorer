{ lib, pkgs, ... }:

{
  options.services.solar_explorer = {
    enable = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Enable the Solar Explorer service.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "solar_explorer_user";
      description = "User to run the Solar Explorer service.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "solar_explorer_group";
      description = "Group to run the Solar Explorer service.";
    };

    configFile = lib.mkOption {
      type = lib.types.path;
      default = "/etc/solar_explorer/config.yml";
      description = "Path to the Solar Explorer configuration file.";
    };
  };

  config = mkIf config.services.solar_explorer.enable {
    systemd.services.solar_explorer = {
      description = "Solar Explorer Service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      serviceConfig = {
        ExecStart = "${pkgs.solar_explorer}/bin/solar_explorer --config ${config.services.solar_explorer.configFile}";
        Restart = "always";
        User = config.services.solar_explorer.user;
        Group = config.services.solar_explorer.group;
      };

      install = {
        wantedBy = [ "multi-user.target" ];
      };
    };
  };
}