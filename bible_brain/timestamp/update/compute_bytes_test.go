package update

/*
func TestGetBoundaries(t *testing.T) {
	ctx := context.Background()
	conn, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatal(status)
	}
	timestamps, status := conn.SelectTimestamps("ENGWEBN2DA", "MRK", 1)
	if status != nil {
		t.Fatal(status)
	}
	directory := filepath.Join(os.Getenv("FCBH_DATASET_FILES"), "ENGWEB", "ENGWEBN2DA")
	filename := filepath.Join(directory, "B02___01_Mark________ENGWEBN2DA.mp3")
	fmt.Println(timestamps[0])
	var segments []Segment
	segments, status = GetBoundaries(ctx, filename, segments)
	if status != nil {
		t.Fatal(status)
	}
	for _, seg := range segments {
		fmt.Println(seg)
	}
	fmt.Println(len(segments))
}
*/
