{ lib, config, pkgs, ... }:
let cfg = config.services.kogs;
in {
  options = with lib; {
    services.kogs = {
      enable = lib.mkEnableOption "Enable kogs";

      listen = mkOption {
        type = types.str;
        default = ":8383";
        description = ''
          Listen string
        '';
      };

      user = mkOption {
        type = with types; oneOf [ str int ];
        default = "kogs";
        description = ''
          The user the service will use.
        '';
      };

      group = mkOption {
        type = with types; oneOf [ str int ];
        default = "kogs";
        description = ''
          The group the service will use.
        '';
      };

      registration = mkOption {
        type = types.bool;
        default = true;
        description = ''
          Allow new users to register
        '';
      };

      dataDir = mkOption {
        type = types.path;
        default = "/var/lib/kogs";
        description = "Path kogs will use to store the database";
      };

      dbDir = mkOption {
        type = types.path;
        default = "${cfg.dataDir}/db";
        description = "Path kogs will use to store the database files";
      };

      package = mkOption {
        type = types.package;
        default = pkgs.kogs;
        defaultText = literalExpression "pkgs.kogs";
        description = "The package to use for kogs";
      };
    };
  };

  config = lib.mkIf (cfg.enable) {
    users.groups.${cfg.group} = { };
    users.users.${cfg.user} = {
      description = "kogs service user";
      isSystemUser = true;
      home = "${cfg.dataDir}";
      createHome = true;
      group = "${cfg.group}";
    };

    systemd.services.kogs = {
      enable = true;
      description = "kogs server";
      wantedBy = [ "network-online.target" ];
      after = [ "network-online.target" ];

      environment = { HOME = "${cfg.dataDir}"; };

      serviceConfig = {
        User = cfg.user;
        Group = cfg.group;

        Restart = "always";

        ExecStart = ''
          ${cfg.package}/bin/kogs -listen ${cfg.listen} -db ${cfg.dbDir} ${lib.optionalString (cfg.registration == false) "-reg=false"}
        '';
      };
    };
  };
}
