{ config, lib, pkgs, solar_exporter, ... }:

with lib;
let
  cfg = config.services.solar_exporter;
in {
  options = {
    services.solar_exporter = {
      enable = mkEnableOption "Solar Exporter service";

      user = mkOption {
        type = types.str;
        default = "solar_exporter";
        description = "User to run the Solar Exporter service.";
      };

      group = mkOption {
        type = types.str;
        default = "solar_exporter";
        description = "Group to run the Solar Exporter service.";
      };

      configFile = mkOption {
        type = types.path;
        default = "/etc/solar_exporter/config.yml";
        description = "Path to the Solar Exporter configuration file.";
      };
    };
  };

  config = mkIf cfg.enable {
    systemd.services.solar_exporter = {
      description = "Solar Exporter Service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      serviceConfig = {
        ExecStart = "${solar_exporter}/bin/solar_exporter --config ${escapeShellArg cfg.configFile}";
        Restart = "always";
        User = cfg.user;
        Group = cfg.group;
      };
    };
  };
}