package lib

import (
	"fmt"
	"testing"
)

func benchmarkReduceToDiff(modulesCount, deltaCount int, b *testing.B) {
	clean()
	defer clean()

	repo, err := createTestRepository(".tmp/repo")
	if err != nil {
		b.Fatalf("%v", err)
	}

	for i := 0; i < modulesCount; i++ {
		err = repo.InitModule(fmt.Sprintf("app-%v", i))
		if err != nil {
			b.Fatalf("%v", err)
		}
	}

	err = repo.Commit("first")
	if err != nil {
		b.Fatalf("%v", err)
	}

	c1 := repo.LastCommit

	for i := 0; i < deltaCount; i++ {
		err = repo.WriteContent(fmt.Sprintf("content/file-%v", i), "sample content")
		if err != nil {
			b.Fatalf("%v", err)
		}
	}

	repo.Commit("second")
	c2 := repo.LastCommit

	world := NewBenchmarkWorld(b, ".tmp/repo")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = world.System.ManifestByDiff(c1.String(), c2.String())
		if err != nil {
			b.Fatalf("%v", err)
		}

	}

	b.StopTimer()
}
func BenchmarkReduceToDiff10(b *testing.B) {
	benchmarkReduceToDiff(10, 10, b)
}

func BenchmarkReduceToDiff100(b *testing.B) {
	benchmarkReduceToDiff(100, 100, b)
}

func BenchmarkReduceToDiff1000(b *testing.B) {
	benchmarkReduceToDiff(1000, 1000, b)
}

func BenchmarkReduceToDiff10000(b *testing.B) {
	benchmarkReduceToDiff(10000, 10000, b)
}
