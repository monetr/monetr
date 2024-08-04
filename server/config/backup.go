package config

type Backup struct {
	// Kind determines what type of destination the backup will use. Accepted
	// values are `s3` and `filesystem`.
	Kind string `yaml:"kind"`
	// Prefix is the prefix for the backup file path that will be used. If your
	// destination is S3 then this will be the folder path to the file in S3. If
	// you are using the filesystem then this should be a path to a folder on the
	// filesystem. The backup file name is not configurable via this field.
	Prefix string `yaml:"path"`
	// ChunkSize specifies how muuch data is "streamed" through memory during the
	// backup process. Larger values will increase memory usage during a backup
	// but may improve performance.
	ChunkSize int `yaml:"chunkSize"`
	// S3 provides the values for the S3 bucket that the backup will be written
	// to. This is only necessary if the kind is S3.
	S3 S3Storage
}
