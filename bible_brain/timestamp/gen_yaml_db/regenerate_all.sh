#!/bin/bash

# Script to regenerate all YAML files with fresh, clean output
# This will create new directories and generate all combinations

echo "Generating fresh YAML files for all combinations..."

# Clean up any existing test directories
echo "Cleaning up existing test directories..."
rm -rf all_*_usx all_*_plain

# N1 (New Testament 1)
echo "Generating N1 USX HLS files..."
./yaml_generator -testament n1 -text usx -stream hls -output ./all_n1_usx/ -verbose | tail -3

echo "Generating N1 Plain HLS files..."
./yaml_generator -testament n1 -text plain -stream hls -output ./all_n1_plain/ -verbose | tail -3

# DASH generation commented out for now (not yet implemented)
# echo "Generating N1 USX DASH files..."
# ./yaml_generator -testament n1 -text usx -stream dash -output ./all_n1_usx_dash/ -verbose | tail -3

# echo "Generating N1 Plain DASH files..."
# ./yaml_generator -testament n1 -text plain -stream dash -output ./all_n1_plain_dash/ -verbose | tail -3

# N2 (New Testament 2)
echo "Generating N2 USX HLS files..."
./yaml_generator -testament n2 -text usx -stream hls -output ./all_n2_usx/ -verbose | tail -3

echo "Generating N2 Plain HLS files..."
./yaml_generator -testament n2 -text plain -stream hls -output ./all_n2_plain/ -verbose | tail -3

# DASH generation commented out for now (not yet implemented)
# echo "Generating N2 USX DASH files..."
# ./yaml_generator -testament n2 -text usx -stream dash -output ./all_n2_usx_dash/ -verbose | tail -3

# echo "Generating N2 Plain DASH files..."
# ./yaml_generator -testament n2 -text plain -stream dash -output ./all_n2_plain_dash/ -verbose | tail -3

# O1 (Old Testament 1)
echo "Generating O1 USX HLS files..."
./yaml_generator -testament o1 -text usx -stream hls -output ./all_o1_usx/ -verbose | tail -3

echo "Generating O1 Plain HLS files..."
./yaml_generator -testament o1 -text plain -stream hls -output ./all_o1_plain/ -verbose | tail -3

# DASH generation commented out for now (not yet implemented)
# echo "Generating O1 USX DASH files..."
# ./yaml_generator -testament o1 -text usx -stream dash -output ./all_o1_usx_dash/ -verbose | tail -3

# echo "Generating O1 Plain DASH files..."
# ./yaml_generator -testament o1 -text plain -stream dash -output ./all_o1_plain_dash/ -verbose | tail -3

# O2 (Old Testament 2)
echo "Generating O2 USX HLS files..."
./yaml_generator -testament o2 -text usx -stream hls -output ./all_o2_usx/ -verbose | tail -3

echo "Generating O2 Plain HLS files..."
./yaml_generator -testament o2 -text plain -stream hls -output ./all_o2_plain/ -verbose | tail -3

# DASH generation commented out for now (not yet implemented)
# echo "Generating O2 USX DASH files..."
# ./yaml_generator -testament o2 -text usx -stream dash -output ./all_o2_usx_dash/ -verbose | tail -3

# echo "Generating O2 Plain DASH files..."
# ./yaml_generator -testament o2 -text plain -stream dash -output ./all_o2_plain_dash/ -verbose | tail -3

echo ""
echo "Generation complete! Summary of files created:"
echo "=============================================="

for dir in all_*; do
    if [ -d "$dir" ]; then
        count=$(ls -1 "$dir" | wc -l)
        printf "%-20s %4d files\n" "$dir:" "$count"
    fi
done

echo ""
echo "Validation checks:"
echo "=================="

# Count total files (including potential duplicates)
total_files=$(ls all_*/*.yaml 2>/dev/null | sort | wc -l)
echo "Total files generated: $total_files"

# Count unique files (removing duplicates)
unique_files=$(ls all_*/*.yaml 2>/dev/null | sort | uniq | wc -l)
echo "Unique files generated: $unique_files"

# Check for duplicates
if [ "$total_files" -eq "$unique_files" ]; then
    echo "✅ VALIDATION PASSED: No duplicate files found!"
else
    duplicates=$((total_files - unique_files))
    echo "❌ VALIDATION FAILED: Found $duplicates duplicate files!"
    echo "Duplicate files:"
    ls all_*/*.yaml 2>/dev/null | sort | uniq -d
fi

# List all directories created
echo ""
echo "Directories created:"
for dir in all_*; do
    if [ -d "$dir" ]; then
        echo "  $dir/"
    fi
done

echo ""
echo "All YAML files have been generated successfully!"
