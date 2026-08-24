package main

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestArchiveOutputFilesArchivesAndRemovesSources(t *testing.T) {
	outputDirectory := t.TempDir()
	projectName := "barley_batch"
	files := []string{
		"final_calls.txt",
		"OUT_FinalReportR.txt",
		"OUT_FinalReportTheta.txt",
		"parms.csv",
		"stats.csv",
		"STATS.txt",
		"ids.csv",
	}

	for _, filename := range files {
		content := []byte("contents of " + filename)
		if err := ioutil.WriteFile(filepath.Join(outputDirectory, filename), content, 0600); err != nil {
			t.Fatalf("write %s: %v", filename, err)
		}
	}

	if err := archiveOutputFilesInDirectory(projectName, outputDirectory); err != nil {
		t.Fatalf("archiveOutputFilesInDirectory() error = %v", err)
	}

	archivePath := filepath.Join(outputDirectory, projectName+".zip")
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer archive.Close()

	var archivedNames []string
	for _, entry := range archive.File {
		archivedNames = append(archivedNames, entry.Name)
		file, err := entry.Open()
		if err != nil {
			t.Fatalf("open archive entry %s: %v", entry.Name, err)
		}
		content, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			t.Fatalf("read archive entry %s: %v", entry.Name, err)
		}
		expected := []byte("contents of " + entry.Name)
		if !reflect.DeepEqual(content, expected) {
			t.Errorf("archive entry %s = %q, want %q", entry.Name, content, expected)
		}
	}
	sort.Strings(archivedNames)
	sortedFiles := append([]string(nil), files...)
	sort.Strings(sortedFiles)
	if !reflect.DeepEqual(archivedNames, sortedFiles) {
		t.Errorf("archive entries = %v, want %v", archivedNames, sortedFiles)
	}

	for _, filename := range files {
		if _, err := os.Stat(filepath.Join(outputDirectory, filename)); !os.IsNotExist(err) {
			t.Errorf("source file %s still exists or returned unexpected error: %v", filename, err)
		}
	}
}

func TestArchiveOutputFilesLeavesSourcesWhenInputIsMissing(t *testing.T) {
	outputDirectory := t.TempDir()
	files := []string{
		"final_calls.txt",
		"OUT_FinalReportR.txt",
		"OUT_FinalReportTheta.txt",
		"parms.csv",
		"stats.csv",
		"ids.csv",
	}
	for _, filename := range files {
		if err := ioutil.WriteFile(filepath.Join(outputDirectory, filename), []byte("data"), 0600); err != nil {
			t.Fatalf("write %s: %v", filename, err)
		}
	}

	if err := archiveOutputFilesInDirectory("incomplete", outputDirectory); err == nil {
		t.Fatal("archiveOutputFilesInDirectory() error = nil, want missing-input error")
	}
	if _, err := os.Stat(filepath.Join(outputDirectory, "incomplete.zip")); !os.IsNotExist(err) {
		t.Fatalf("incomplete archive exists or returned unexpected error: %v", err)
	}
	for _, filename := range files {
		if _, err := os.Stat(filepath.Join(outputDirectory, filename)); err != nil {
			t.Errorf("source file %s was not preserved: %v", filename, err)
		}
	}
}

func TestProcessGenotypeDataMapsHeaderAndWritesStats(t *testing.T) {
	workingDirectory := t.TempDir()
	outputDirectory := filepath.Join(workingDirectory, "sample_data")
	if err := os.Mkdir(outputDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(outputDirectory, "final_calls.txt"), []byte("Line/Marker\tRow.Names\tUNKNOWN\tSENTRIX_1\nmarker_b\tAA\tN/A\tCA\nmarker_a\tCC\tN/A\tAG\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(workingDirectory, "20170301_44040_ALL_NAMES_MAP.txt"), []byte("Lines/Markers\tLines/Markers\nmarker_a\tMappedA\nmarker_b\tMappedB\n"), 0600); err != nil {
		t.Fatal(err)
	}

	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workingDirectory); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWorkingDirectory)

	oldProject := *projectFilePtr
	*projectFilePtr = filepath.Join(workingDirectory, "project_alpha")
	defer func() { *projectFilePtr = oldProject }()

	ProcessGenotypeData(map[string]string{"SENTRIX_1": "Sample_001"})

	outputPath := filepath.Join(outputDirectory, "project_alpha_FINAL.txt")
	output, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read processed output: %v", err)
	}
	expectedOutput := "Line/Marker\t\tSample_001\tUNKNOWN\nMappedA\tC\tA/G\t--\nMappedB\tA\tA/C\t--\n"
	if string(output) != expectedOutput {
		t.Errorf("processed output = %q, want %q", output, expectedOutput)
	}

	stats, err := ioutil.ReadFile(filepath.Join(outputDirectory, "STATS.txt"))
	if err != nil {
		t.Fatalf("read stats: %v", err)
	}
	expectedStats := "--\t2\nA\t1\nA/C\t1\nA/G\t1\nC\t1\n"
	if string(stats) != expectedStats {
		t.Errorf("stats = %q, want %q", stats, expectedStats)
	}
}
