import sbt._
import Keys._
import skinny.servlet.ServletPlugin._
import skinny.servlet.ServletKeys._
import com.mojolly.scalate.ScalatePlugin._
import ScalateKeys._
import com.typesafe.config.ConfigFactory

lazy val skinnyMicroVersion = "1.1.0"
lazy val jettyVersion = "9.3.11.v20160721"

lazy val commonSettings = Seq(
  scalaVersion in Global := "2.11.8",
  scalacOptions ++= Seq(
    "-unchecked",
    "-deprecation",
    "-feature",
    "-encoding", "UTF-8"
  ),
  javaOptions ++= Seq(
    "-Dfile.encoding=UTF-8"
  )
)

def configFile(project: String): java.io.File =
  file(project) / "src" / "main" / "resources" / "application.conf"

lazy val isucon = (project in file(".")).
  settings(commonSettings: _*)

lazy val isuda = (project in file("isuda")).
  settings(commonSettings: _*).
  settings(servletSettings: _*).
  settings(scalateSettings: _*).
  settings(
    name := "isuda",
    version := "0.0.1",
    libraryDependencies ++= Seq(
      "org.skinny-framework" %% "skinny-micro" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-micro-json4s" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-micro-scalate" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-micro-server" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-http-client" % "2.2.0",
      "org.eclipse.jetty" % "jetty-webapp" % jettyVersion % "container",
      "org.eclipse.jetty" % "jetty-plus"   % jettyVersion % "container",
      "org.scalikejdbc" %% "scalikejdbc" % "2.4.2",
      "org.scalikejdbc" %% "scalikejdbc-config" % "2.4.2",
      "mysql" % "mysql-connector-java" % "6.0.3",
      "ch.qos.logback" % "logback-classic" % "1.1.7"
    ),

    // Local server
    port in container.Configuration := isudaConfig.getInt("server.port"),

    // Static files
    unmanagedResourceDirectories in Compile +=
      baseDirectory(_ / ".." / ".." / "public").value,

    // Templates
    scalateTemplateConfig in Compile <<= (sourceDirectory in Compile) { base =>
      Seq(
        TemplateConfig(
          base / "webapp" / "WEB-INF" / "views",
          Seq.empty,
          Seq.empty,
          Some("views")
        ),
        TemplateConfig(
          base / "webapp" / "WEB-INF" / "templates",
          Seq.empty,
          Seq.empty,
          Some("templates")
        )
      )
    },
    unmanagedResourceDirectories in Compile ++=
      (webappResources in Compile).value,

    // Packaging
    mainClass in assembly := Some("skinny.standalone.JettyLauncher"),
    sbt.Keys.test in assembly := {},
    assemblyOutputPath in assembly := file(".") / "isuda.jar"
  )

def isudaConfig: com.typesafe.config.Config =
  ConfigFactory.parseFile(configFile("isuda")).resolve

lazy val isutar = (project in file("isutar")).
  settings(commonSettings: _*).
  settings(servletSettings: _*).
  settings(
    name := "isutar",
    version := "0.0.1",
    libraryDependencies ++= Seq(
      "org.skinny-framework" %% "skinny-micro" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-micro-json4s" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-micro-server" % skinnyMicroVersion,
      "org.skinny-framework" %% "skinny-http-client" % "2.2.0",
      "org.eclipse.jetty" % "jetty-webapp" % jettyVersion % "container",
      "org.eclipse.jetty" % "jetty-plus"   % jettyVersion % "container",
      "org.scalikejdbc" %% "scalikejdbc" % "2.4.2",
      "org.scalikejdbc" %% "scalikejdbc-config" % "2.4.2",
      "mysql" % "mysql-connector-java" % "6.0.3",
      "ch.qos.logback" % "logback-classic" % "1.1.7"
    ),
    port in container.Configuration := isutarConfig.getInt("server.port"),
    mainClass in assembly := Some("skinny.standalone.JettyLauncher"),
    sbt.Keys.test in assembly := {},
    assemblyOutputPath in assembly := file(".") / "isutar.jar"
  )

def isutarConfig: com.typesafe.config.Config =
  ConfigFactory.parseFile(configFile("isutar")).resolve
