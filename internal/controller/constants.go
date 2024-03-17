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

	DefaultClusterDomainName = "clusterDomain"
	DefaultClusterDomain     = "cluster.local"

	DefaultLogVolumeName = "log"

	DefaultConfigNameSuffix = "-config"

	DefaultHome              = "/opt/seatunnel"
	DefaultConfigFileName    = "seatunnel.conf"
	DefaultLogConfigFileName = "log4j.properties"
	DefaultDiskPathPrefix    = "disk"

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
