{ lib, pkgs, config, ... }:

with lib;                      
let
  cfg = config.services.solar_exporter;
in {
  options = {
    services.solar_exporter = {
      enable = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = "Enable the Solar Exporter service.";
      };

      user = lib.mkOption {
        type = lib.types.str;
        default = "solar_exporter_user";
        description = "User to run the Solar Exporter service.";
      };

      group = lib.mkOption {
        type = lib.types.str;
        default = "solar_exporter_group";
        description = "Group to run the Solar Exporter service.";
      };

      configFile = lib.mkOption {
        type = lib.types.path;
        default = "/etc/solar_exporter/config.yml";
        description = "Path to the Solar Exporter configuration file.";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.solar_exporter = {
      description = "Solar Exporter Service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      serviceConfig = {
        ExecStart = "${pkgs.solar_exporter}/bin/solar_exporter --config ${escapeShellArg cfg.configFile}";
        Restart = "always";
        User = cfg.user;
        Group = cfg.group;
      };
    };
  };
}
