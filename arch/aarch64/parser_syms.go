package aarch64

var _Symbols = map[string]interface{} {
    "EQ"              : EQ,
    "NE"              : NE,
    "CS"              : CS,
    "CC"              : CC,
    "MI"              : MI,
    "PL"              : PL,
    "VS"              : VS,
    "VC"              : VC,
    "HI"              : HI,
    "LS"              : LS,
    "GE"              : GE,
    "LT"              : LT,
    "GT"              : GT,
    "LE"              : LE,
    "AL"              : AL,
    "HS"              : HS,
    "LO"              : LO,
    "SY"              : SY,
    "ST"              : ST,
    "LD"              : LD,
    "ISH"             : ISH,
    "ISHST"           : ISHST,
    "ISHLD"           : ISHLD,
    "NSH"             : NSH,
    "NSHST"           : NSHST,
    "NSHLD"           : NSHLD,
    "OSH"             : OSH,
    "OSHST"           : OSHST,
    "OSHLD"           : OSHLD,
    "S1E1R"           : S1E1R,
    "S1E1W"           : S1E1W,
    "S1E0R"           : S1E0R,
    "S1E0W"           : S1E0W,
    "S1E1RP"          : S1E1RP,
    "S1E1WP"          : S1E1WP,
    "S1E2R"           : S1E2R,
    "S1E2W"           : S1E2W,
    "S12E1R"          : S12E1R,
    "S12E1W"          : S12E1W,
    "S12E0R"          : S12E0R,
    "S12E0W"          : S12E0W,
    "S1E3R"           : S1E3R,
    "S1E3W"           : S1E3W,
    "IALL"            : IALL,
    "INJ"             : INJ,
    "IVAC"            : IVAC,
    "ISW"             : ISW,
    "IGVAC"           : IGVAC,
    "IGSW"            : IGSW,
    "IGDVAC"          : IGDVAC,
    "IGDSW"           : IGDSW,
    "CSW"             : CSW,
    "CGSW"            : CGSW,
    "CGDSW"           : CGDSW,
    "CISW"            : CISW,
    "CIGSW"           : CIGSW,
    "CIGDSW"          : CIGDSW,
    "ZVA"             : ZVA,
    "GVA"             : GVA,
    "GZVA"            : GZVA,
    "CVAC"            : CVAC,
    "CGVAC"           : CGVAC,
    "CGDVAC"          : CGDVAC,
    "CVAU"            : CVAU,
    "CVAP"            : CVAP,
    "CGVAP"           : CGVAP,
    "CGDVAP"          : CGDVAP,
    "CVADP"           : CVADP,
    "CGVADP"          : CGVADP,
    "CGDVADP"         : CGDVADP,
    "CIVAC"           : CIVAC,
    "CIGVAC"          : CIGVAC,
    "CIGDVAC"         : CIGDVAC,
    "CIPAE"           : CIPAE,
    "CIGDPAE"         : CIGDPAE,
    "CIPAPA"          : CIPAPA,
    "CIGDPAPA"        : CIGDPAPA,
    "IALLUIS"         : IALLUIS,
    "IALLU"           : IALLU,
    "IVAU"            : IVAU,
    "VMALLE1OS"       : VMALLE1OS,
    "VAE1OS"          : VAE1OS,
    "ASIDE1OS"        : ASIDE1OS,
    "VAAE1OS"         : VAAE1OS,
    "VALE1OS"         : VALE1OS,
    "VAALE1OS"        : VAALE1OS,
    "RVAE1IS"         : RVAE1IS,
    "RVAAE1IS"        : RVAAE1IS,
    "RVALE1IS"        : RVALE1IS,
    "RVAALE1IS"       : RVAALE1IS,
    "VMALLE1IS"       : VMALLE1IS,
    "VAE1IS"          : VAE1IS,
    "ASIDE1IS"        : ASIDE1IS,
    "VAAE1IS"         : VAAE1IS,
    "VALE1IS"         : VALE1IS,
    "VAALE1IS"        : VAALE1IS,
    "RVAE1OS"         : RVAE1OS,
    "RVAAE1OS"        : RVAAE1OS,
    "RVALE1OS"        : RVALE1OS,
    "RVAALE1OS"       : RVAALE1OS,
    "RVAE1"           : RVAE1,
    "RVAAE1"          : RVAAE1,
    "RVALE1"          : RVALE1,
    "RVAALE1"         : RVAALE1,
    "VMALLE1"         : VMALLE1,
    "VAE1"            : VAE1,
    "ASIDE1"          : ASIDE1,
    "VAAE1"           : VAAE1,
    "VALE1"           : VALE1,
    "VAALE1"          : VAALE1,
    "VMALLE1OSNXS"    : VMALLE1OSNXS,
    "VAE1OSNXS"       : VAE1OSNXS,
    "ASIDE1OSNXS"     : ASIDE1OSNXS,
    "VAAE1OSNXS"      : VAAE1OSNXS,
    "VALE1OSNXS"      : VALE1OSNXS,
    "VAALE1OSNXS"     : VAALE1OSNXS,
    "RVAE1ISNXS"      : RVAE1ISNXS,
    "RVAAE1ISNXS"     : RVAAE1ISNXS,
    "RVALE1ISNXS"     : RVALE1ISNXS,
    "RVAALE1ISNXS"    : RVAALE1ISNXS,
    "VMALLE1ISNXS"    : VMALLE1ISNXS,
    "VAE1ISNXS"       : VAE1ISNXS,
    "ASIDE1ISNXS"     : ASIDE1ISNXS,
    "VAAE1ISNXS"      : VAAE1ISNXS,
    "VALE1ISNXS"      : VALE1ISNXS,
    "VAALE1ISNXS"     : VAALE1ISNXS,
    "RVAE1OSNXS"      : RVAE1OSNXS,
    "RVAAE1OSNXS"     : RVAAE1OSNXS,
    "RVALE1OSNXS"     : RVALE1OSNXS,
    "RVAALE1OSNXS"    : RVAALE1OSNXS,
    "RVAE1NXS"        : RVAE1NXS,
    "RVAAE1NXS"       : RVAAE1NXS,
    "RVALE1NXS"       : RVALE1NXS,
    "RVAALE1NXS"      : RVAALE1NXS,
    "VMALLE1NXS"      : VMALLE1NXS,
    "VAE1NXS"         : VAE1NXS,
    "ASIDE1NXS"       : ASIDE1NXS,
    "VAAE1NXS"        : VAAE1NXS,
    "VALE1NXS"        : VALE1NXS,
    "VAALE1NXS"       : VAALE1NXS,
    "IPAS2E1IS"       : IPAS2E1IS,
    "RIPAS2E1IS"      : RIPAS2E1IS,
    "IPAS2LE1IS"      : IPAS2LE1IS,
    "RIPAS2LE1IS"     : RIPAS2LE1IS,
    "ALLE2OS"         : ALLE2OS,
    "VAE2OS"          : VAE2OS,
    "ALLE1OS"         : ALLE1OS,
    "VALE2OS"         : VALE2OS,
    "VMALLS12E1OS"    : VMALLS12E1OS,
    "RVAE2IS"         : RVAE2IS,
    "RVALE2IS"        : RVALE2IS,
    "ALLE2IS"         : ALLE2IS,
    "VAE2IS"          : VAE2IS,
    "ALLE1IS"         : ALLE1IS,
    "VALE2IS"         : VALE2IS,
    "VMALLS12E1IS"    : VMALLS12E1IS,
    "IPAS2E1OS"       : IPAS2E1OS,
    "IPAS2E1"         : IPAS2E1,
    "RIPAS2E1"        : RIPAS2E1,
    "RIPAS2E1OS"      : RIPAS2E1OS,
    "IPAS2LE1OS"      : IPAS2LE1OS,
    "IPAS2LE1"        : IPAS2LE1,
    "RIPAS2LE1"       : RIPAS2LE1,
    "RIPAS2LE1OS"     : RIPAS2LE1OS,
    "RVAE2OS"         : RVAE2OS,
    "RVALE2OS"        : RVALE2OS,
    "RVAE2"           : RVAE2,
    "RVALE2"          : RVALE2,
    "ALLE2"           : ALLE2,
    "VAE2"            : VAE2,
    "ALLE1"           : ALLE1,
    "VALE2"           : VALE2,
    "VMALLS12E1"      : VMALLS12E1,
    "IPAS2E1ISNXS"    : IPAS2E1ISNXS,
    "RIPAS2E1ISNXS"   : RIPAS2E1ISNXS,
    "IPAS2LE1ISNXS"   : IPAS2LE1ISNXS,
    "RIPAS2LE1ISNXS"  : RIPAS2LE1ISNXS,
    "ALLE2OSNXS"      : ALLE2OSNXS,
    "VAE2OSNXS"       : VAE2OSNXS,
    "ALLE1OSNXS"      : ALLE1OSNXS,
    "VALE2OSNXS"      : VALE2OSNXS,
    "VMALLS12E1OSNXS" : VMALLS12E1OSNXS,
    "RVAE2ISNXS"      : RVAE2ISNXS,
    "RVALE2ISNXS"     : RVALE2ISNXS,
    "ALLE2ISNXS"      : ALLE2ISNXS,
    "VAE2ISNXS"       : VAE2ISNXS,
    "ALLE1ISNXS"      : ALLE1ISNXS,
    "VALE2ISNXS"      : VALE2ISNXS,
    "VMALLS12E1ISNXS" : VMALLS12E1ISNXS,
    "IPAS2E1OSNXS"    : IPAS2E1OSNXS,
    "IPAS2E1NXS"      : IPAS2E1NXS,
    "RIPAS2E1NXS"     : RIPAS2E1NXS,
    "RIPAS2E1OSNXS"   : RIPAS2E1OSNXS,
    "IPAS2LE1OSNXS"   : IPAS2LE1OSNXS,
    "IPAS2LE1NXS"     : IPAS2LE1NXS,
    "RIPAS2LE1NXS"    : RIPAS2LE1NXS,
    "RIPAS2LE1OSNXS"  : RIPAS2LE1OSNXS,
    "RVAE2OSNXS"      : RVAE2OSNXS,
    "RVALE2OSNXS"     : RVALE2OSNXS,
    "RVAE2NXS"        : RVAE2NXS,
    "RVALE2NXS"       : RVALE2NXS,
    "ALLE2NXS"        : ALLE2NXS,
    "VAE2NXS"         : VAE2NXS,
    "ALLE1NXS"        : ALLE1NXS,
    "VALE2NXS"        : VALE2NXS,
    "VMALLS12E1NXS"   : VMALLS12E1NXS,
    "ALLE3OS"         : ALLE3OS,
    "VAE3OS"          : VAE3OS,
    "PAALLOS"         : PAALLOS,
    "VALE3OS"         : VALE3OS,
    "RVAE3IS"         : RVAE3IS,
    "RVALE3IS"        : RVALE3IS,
    "ALLE3IS"         : ALLE3IS,
    "VAE3IS"          : VAE3IS,
    "VALE3IS"         : VALE3IS,
    "RPAOS"           : RPAOS,
    "RPALOS"          : RPALOS,
    "RVAE3OS"         : RVAE3OS,
    "RVALE3OS"        : RVALE3OS,
    "RVAE3"           : RVAE3,
    "RVALE3"          : RVALE3,
    "ALLE3"           : ALLE3,
    "VAE3"            : VAE3,
    "PAALL"           : PAALL,
    "VALE3"           : VALE3,
    "ALLE3OSNXS"      : ALLE3OSNXS,
    "VAE3OSNXS"       : VAE3OSNXS,
    "VALE3OSNXS"      : VALE3OSNXS,
    "RVAE3ISNXS"      : RVAE3ISNXS,
    "RVALE3ISNXS"     : RVALE3ISNXS,
    "ALLE3ISNXS"      : ALLE3ISNXS,
    "VAE3ISNXS"       : VAE3ISNXS,
    "VALE3ISNXS"      : VALE3ISNXS,
    "RVAE3OSNXS"      : RVAE3OSNXS,
    "RVALE3OSNXS"     : RVALE3OSNXS,
    "RVAE3NXS"        : RVAE3NXS,
    "RVALE3NXS"       : RVALE3NXS,
    "ALLE3NXS"        : ALLE3NXS,
    "VAE3NXS"         : VAE3NXS,
    "VALE3NXS"        : VALE3NXS,
    "SM"              : SM,
    "ZA"              : ZA,
    "CSYNC"           : CSYNC,
    "DSYNC"           : DSYNC,
    "RCTX"            : RCTX,
    "PLD_L1_KEEP"     : PLD_L1_KEEP,
    "PLD_L1_STRM"     : PLD_L1_STRM,
    "PLD_L2_KEEP"     : PLD_L2_KEEP,
    "PLD_L2_STRM"     : PLD_L2_STRM,
    "PLD_L3_KEEP"     : PLD_L3_KEEP,
    "PLD_L3_STRM"     : PLD_L3_STRM,
    "PLI_L1_KEEP"     : PLI_L1_KEEP,
    "PLI_L1_STRM"     : PLI_L1_STRM,
    "PLI_L2_KEEP"     : PLI_L2_KEEP,
    "PLI_L2_STRM"     : PLI_L2_STRM,
    "PLI_L3_KEEP"     : PLI_L3_KEEP,
    "PLI_L3_STRM"     : PLI_L3_STRM,
    "PST_L1_KEEP"     : PST_L1_KEEP,
    "PST_L1_STRM"     : PST_L1_STRM,
    "PST_L2_KEEP"     : PST_L2_KEEP,
    "PST_L2_STRM"     : PST_L2_STRM,
    "PST_L3_KEEP"     : PST_L3_KEEP,
    "PST_L3_STRM"     : PST_L3_STRM,
    "PLD_KEEP"        : PLD_KEEP,
    "PLD_STRM"        : PLD_STRM,
    "PST_KEEP"        : PST_KEEP,
    "PST_STRM"        : PST_STRM,
}
