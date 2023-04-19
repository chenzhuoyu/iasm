<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                version="1.0" xmlns="http://www.w3.org/1999/xhtml">
<xsl:output doctype-public="-//W3C//DTD XHTML 1.1//EN"
 doctype-system="http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"
 method="html" encoding="utf-8"
 indent="yes"/>

  <xsl:template match="/register_index">
    <html>
      <xsl:apply-templates />
    </html>
  </xsl:template>

  <xsl:template match="/register_index/toptitle">
    <head>
      <title><xsl:value-of select="@architecture"/></title>
      <link rel="stylesheet" type="text/css" href="insn.css"/>
    </head>
  </xsl:template>

  <xsl:template match="/register_index/register_links">
    <body>
      <xsl:call-template name="generate_header_footer"/>
      <hr/>
      <h1 class="alphindextitle"><xsl:value-of select="@title"/></h1>
      <xsl:apply-templates />
      <hr/>
      <xsl:call-template name="generate_header_footer"/>
      <p class="versions"><xsl:value-of select="../timestamp"/></p>
      <p class="copyconf">Copyright &#169; 2010-2023 Arm Limited or its affiliates. All rights reserved. This document is Non-Confidential.</p>
     </body>
  </xsl:template>

  <xsl:template name="generate_header_footer">
    <table style="margin: 0 auto;">
      <tr>
        <!-- header/footer start -->
        <td><div class="topbar"><a href="AArch32-regindex.xml">AArch32 Registers</a></div></td>
        <td><div class="topbar"><a href="AArch64-regindex.xml">AArch64 Registers</a></div></td>
        <td><div class="topbar"><a href="AArch32-sysindex.xml">AArch32 Instructions</a></div></td>
        <td><div class="topbar"><a href="AArch64-sysindex.xml">AArch64 Instructions</a></div></td>
        <td><div class="topbar"><a href="enc_index.xml">Index by Encoding</a></div></td>
        <td><div class="topbar"><a href="ext_alpha_index.xml">External Registers</a></div></td>
        <td><div class="topbar"><a href="ext_enc_index.xml">External Registers by Offset</a></div></td>
        <td><div class="topbar"><a href="func_index.xml">Registers by Functional Group</a></div></td>
        <td><div class="topbar"><a href="notice.xml">Proprietary Notice</a></div></td>
        <!-- header/footer end -->
      </tr>
    </table>
  </xsl:template>

  <xsl:template match="/register_index/register_links/register_link">
  <div>
    <p class="iformindex">
      <span class="insnheading">
        <a>
          <xsl:attribute name="href">
            <xsl:choose>
              <xsl:when test="@id = ''">
                <xsl:value-of select="@href"/>
              </xsl:when>
              <xsl:otherwise>
                <xsl:value-of select="@registerfile"/>
              </xsl:otherwise>
            </xsl:choose>
          </xsl:attribute>
          <xsl:value-of select="@heading"/>
        </a>:
        <xsl:value-of select="."/>
      </span>
    </p>
  </div>
  </xsl:template>
  
  <!-- We don't need a separate timestamp template, as we've already sorted it above.-->
  <xsl:template match="/register_index/timestamp">
  </xsl:template>

</xsl:stylesheet>
