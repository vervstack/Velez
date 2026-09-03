package vervonomicon

// Box is a named sizing tier, stored in velez.resource_boxes. An
// administrator may edit or extend the seeded builtin tiers with custom
// ones.
type Box struct {
	Name      string
	Cpu       float64
	RamMb     int64
	DiskMb    int64
	IsBuiltin bool
}
