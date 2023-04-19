<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                version="1.0" xmlns="http://www.w3.org/1999/xhtml">
<xsl:output doctype-public="-//W3C//DTD XHTML 1.1//EN"
 doctype-system="http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"
 method="html" encoding="utf-8"
 indent="yes"/>

  <xsl:template match="/sysregindex">
    <html>
      <head>
        <title>
	        <xsl:choose>
	          <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'True'">
	            System Register index by encoding
	          </xsl:when>
	          <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'False'">
              External register index by offset
            </xsl:when>
            <xsl:when test="/sysregindex/@indextype = 'function'">
              Index by functional group
            </xsl:when>
	        </xsl:choose>
        </title>
        <link rel="stylesheet" type="text/css" href="insn.css"/>
      </head>
      <body>
        <xsl:call-template name="generate_header_footer"/>
        <hr/>
        <xsl:choose>
          <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'True'">
           <h1 class="sysregindex">System Register index by instruction and encoding</h1>
          </xsl:when>
          <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'False'">
            <h1 class="sysregindex">External register index by offset</h1>
          </xsl:when>
          <xsl:when test="/sysregindex/@indextype = 'function'">
            <h1 class="sysregindex">Index by functional group</h1>
          </xsl:when>
        </xsl:choose>

        <xsl:apply-templates select="typelist"/>
        <xsl:apply-templates select="sectiongroup"/>
        <hr/>
        <xsl:call-template name="generate_header_footer"/>
        <p class="versions"><xsl:value-of select="timestamp"/></p>
        <p class="copyconf">Copyright &#169; 2010-2023 Arm Limited or its affiliates. All rights reserved. This document is Non-Confidential.</p>
      </body>
    </html>
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

  <xsl:template match="typelist">
    <xsl:choose>
      <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'True'">
        <p>Below are indexes for registers and operations accessed in the following ways:</p>
      </xsl:when>
      <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'False'">
        <p>Below are indexes for external registers in the following blocks:</p>
      </xsl:when>
      <xsl:when test="/sysregindex/@indextype = 'function'">
        <p>Below are indexes for registers with the following main functional groups:</p>
      </xsl:when>
    </xsl:choose>
    <xsl:for-each select="typegroup">
      <xsl:if test="/sysregindex/@is_internal = 'True'"><p>For <xsl:value-of select="@groupname"/></p></xsl:if>
      <ul>
        <xsl:for-each select="typelink">
          <li>
            <a href="#{@link}"><xsl:value-of select="@type"/></a>
          </li>
        </xsl:for-each>
      </ul>
    </xsl:for-each>
  </xsl:template>
  
  <xsl:template match="sysregindex/sectiongroup">
    <xsl:if test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'True'"><h2>Registers and operations in <xsl:value-of select="@groupname"/></h2></xsl:if>
    <xsl:apply-templates/>
  </xsl:template>

  <xsl:template match="section">
    <h2 class="sysregindex">
      <a id="{@anchor}">
		    <xsl:choose>
		      <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'True'">
		        Accessed using <xsl:value-of select="@type"/>:
		      </xsl:when>
		      <xsl:when test="/sysregindex/@indextype = 'encoding' and /sysregindex/@is_internal = 'False'">
		        In the <xsl:value-of select="@type"/> block:
		      </xsl:when>
		      <xsl:when test="/sysregindex/@indextype = 'function'">
		        In the <xsl:value-of select="@type"/> functional group:
		      </xsl:when>
		    </xsl:choose>
      </a>
    </h2>
    <table class="instructiontable">
      <xsl:apply-templates />
    </table>
  </xsl:template>

  <xsl:template match="heading">
    <thead>
      <xsl:apply-templates />
    </thead>
  </xsl:template>

  <xsl:template match="tbody">
    <tbody>
      <xsl:apply-templates />
    </tbody>
  </xsl:template>

  <xsl:template match="row">
    <tr>
      <xsl:if test="@class">
        <xsl:attribute name="class">
          <xsl:value-of select="@class"/>
        </xsl:attribute>
      </xsl:if>
      <xsl:apply-templates />
    </tr>
  </xsl:template>

  <xsl:template match="heading/row/entry">
    <th>
      <xsl:if test="@colspan">
        <xsl:attribute name="colspan">
          <xsl:value-of select="@colspan"/>
        </xsl:attribute>
      </xsl:if>
      <xsl:if test="@rowspan">
        <xsl:attribute name="rowspan">
          <xsl:value-of select="@rowspan"/>
        </xsl:attribute>
      </xsl:if>
      <xsl:if test="../@class = 'header2'">
        <xsl:attribute name="class">bitfields</xsl:attribute>
      </xsl:if>
      <xsl:value-of select="."/>
    </th>
  </xsl:template>

  <xsl:template match="tbody/row/entry">
    <td>
      <xsl:if test="@class">
        <xsl:attribute name="class">
          <xsl:value-of select="@class"/>
        </xsl:attribute>
      </xsl:if>
      <xsl:apply-templates/>
    </td>
  </xsl:template>

  <xsl:template match="a">
    <a href="{@href}"><xsl:value-of select="."/></a>
  </xsl:template>

</xsl:stylesheet>
