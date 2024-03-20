package controller

import (
	"fmt"
)

const (
	// DefaultNameSuffix is the default name suffix of the resources of the kafka
	DefaultNameSuffix = "-seatunnel"

	// DefaultClusterSign is the default cluster sign of the kafka
	DefaultClusterSign = "seatunnel"

	// DefaultStorageClass is the default storage class of the kafka
	DefaultStorageClass = "nineinfra-default"

	// DefaultReplicas is the default replicas
	DefaultReplicas = 3

	// DefaultDiskNum is the default disk num
	DefaultDiskNum = 1

	// DefaultClusterDomainName is the cluster domain name
	DefaultClusterDomainName = "clusterDomain"

	// DefaultClusterDomain is the cluster domain value
	DefaultClusterDomain = "cluster.local"

	// DefaultLogVolumeName is the log volume name
	DefaultLogVolumeName = "log"

	// DefaultConfigNameSuffix is the config volume name suffix
	DefaultConfigNameSuffix = "-config"

	// DefaultClusterRefsConfigNameSuffix is the cluster refs config volume name suffix
	DefaultClusterRefsConfigNameSuffix = "-clusterrefs"

	// DefaultHome is the home dir
	DefaultHome = "/opt/seatunnel"

	// DefaultConfigFileName is the config file name
	DefaultConfigFileName = "seatunnel.conf"

	// DefaultLogConfigFileName is the log config file name
	DefaultLogConfigFileName = "log4j.properties"

	// DefaultDiskPathPrefix is disk path prefix
	DefaultDiskPathPrefix = "disk"

	// DefaultSparkConfFile is the default conf file name of the spark
	DefaultSparkConfFile = "spark-defaults.conf"

	// DefaultSparkConfDir is the default conf dir of the spark
	DefaultSparkConfDir = "/opt/spark/conf"

	// DefaultHiveSiteFile is the default hive site file name of the metastore
	DefaultHiveSiteFile = "hive-site.xml"

	// DefaultHdfsSiteFile is the default hdfs site file name of the hdfs
	DefaultHdfsSiteFile = "hdfs-site.xml"

	// DefaultCoreSiteFile is the default core site file name of the hdfs
	DefaultCoreSiteFile = "core-site.xml"

	// DefaultTerminationGracePeriod is the default time given before the
	// container is stopped. This gives clients time to disconnect from a
	// specific node gracefully.
	DefaultTerminationGracePeriod = 30

	// DefaultVolumeSize is the default size for the volume
	DefaultVolumeSize    = "50Gi"
	DefaultLogVolumeSize = "5Gi"

	// DefaultReadinessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the readiness probe
	DefaultReadinessProbeInitialDelaySeconds = 40

	// DefaultReadinessProbePeriodSeconds is the default probe period (in seconds)
	// for the readiness probe
	DefaultReadinessProbePeriodSeconds = 10

	// DefaultReadinessProbeFailureThreshold is the default probe failure threshold
	// for the readiness probe
	DefaultReadinessProbeFailureThreshold = 10

	// DefaultReadinessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultReadinessProbeSuccessThreshold = 1

	// DefaultReadinessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the readiness probe
	DefaultReadinessProbeTimeoutSeconds = 10

	// DefaultLivenessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the liveness probe
	DefaultLivenessProbeInitialDelaySeconds = 40

	// DefaultLivenessProbePeriodSeconds is the default probe period (in seconds)
	// for the liveness probe
	DefaultLivenessProbePeriodSeconds = 10

	// DefaultLivenessProbeFailureThreshold is the default probe failure threshold
	// for the liveness probe
	DefaultLivenessProbeFailureThreshold = 10

	// DefaultLivenessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultLivenessProbeSuccessThreshold = 1

	// DefaultLivenessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the liveness probe
	DefaultLivenessProbeTimeoutSeconds = 10

	//DefaultProbeTypeLiveness liveness type probe
	DefaultProbeTypeLiveness = "liveness"

	//DefaultProbeTypeReadiness readiness type probe
	DefaultProbeTypeReadiness = "readiness"
)

var (
	DefaultConfPath = fmt.Sprintf("%s/%s", DefaultHome, "conf")
	DefaultDataPath = fmt.Sprintf("%s/%s", DefaultHome, "data")
	DefaultLogPath  = fmt.Sprintf("%s/%s", DefaultHome, "logs")
	DefaultConfFile = fmt.Sprintf("%s/%s", DefaultConfPath, DefaultConfigFileName)
	ClusterRefsConfFileList = []string{DefaultSparkConfFile, DefaultHiveSiteFile, DefaultHdfsSiteFile, DefaultCoreSiteFile}
)

var (
	SourceTypeSupportedList    = "Kafka,Jdbc,HdfsFile,S3File,MongoDB,MySQL-CDC,Postgres-CDC,SqlServer-CDC,Oracle-CDC,MongoDB-CDC"
	TransformTypeSupportedList = "Sql,Split,Replace,JsonPath,Copy,FieldMapper,FilterRowKind,Filter"
	SinkTypeSupportedList      = "Kafka,Jdbc,HdfsFile,S3File,MongoDB,Doris,Clickhouse"
)

var DefaultLogConfKeyValue = map[string]string{
	"log4j.rootLogger":                                "INFO, Console, RFA",
	"log4j.appender.Console":                          "org.apache.log4j.ConsoleAppender",
	"log4j.appender.Console.layout":                   "org.apache.log4j.PatternLayout",
	"log4j.appender.Console.layout.ConversionPattern": "[%d] %p %m (%c)%n",

	"log4j.appender.RFA":                          "org.apache.log4j.DailyRollingFileAppender",
	"log4j.appender.RFA.DatePattern":              "'.'yyyy-MM-dd-HH",
	"log4j.appender.RFA.File":                     "${kafka.logs.dir}/server.log",
	"log4j.appender.RFA.layout":                   "org.apache.log4j.PatternLayout",
	"log4j.appender.RFA.layout.ConversionPattern": "[%d] %p %m (%c)%n",
}
