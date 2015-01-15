package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/docker/docker/pkg/archive"
)

const REPO_PATH string = "/tmp/git_repo"

var (
	branches []string = []string{
		"layer_0_4986bf8c15363d1c5d15512d5266f8777bfba4974ac56e3270e7760f6f0a8125",
		"layer_1_4986bf8c15363d1c5d15512d5266f8777bfba4974ac56e3270e7760f6f0a8126",
		"layer_2_4986bf8c15363d1c5d15512d5266f8777bfba4974ac56e3270e7760f6f0a8127",
		"layer_3_4986bf8c15363d1c5d15512d5266f8777bfba4974ac56e3270e7760f6f0a8128",
	}
)

func assertErrNil(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestGitFlow(t *testing.T) {
	fmt.Printf("Testing git ... ")
	r, err := newGitRepo(REPO_PATH)
	assertErrNil(err, t)

	defer os.RemoveAll(REPO_PATH)

	//Create 3 branches
	for i := 0; i < 3; i++ {
		br := branches[i]
		_, err = r.checkoutB(br)
		assertErrNil(err, t)

		curBr, err := r.currentBranch()
		assertErrNil(err, t)

		if br != curBr {
			t.Fatalf("current branch: %v expected %v", curBr, br)
		}

		f, err := os.Create(path.Join(r.Path, "br"+strconv.Itoa(i)+".txt"))
		assertErrNil(err, t)
		f.Close()

		_, err = r.addAllAndCommit("commit message")
		assertErrNil(err, t)
	}

	exportChangeSet(r, branches[0], []string{"br0.txt"}, []string{"br1.txt", "br2.txt", ".git"}, t)
	exportChangeSet(r, branches[1], []string{"br1.txt"}, []string{"br0.txt", "br2.txt"}, t)
	exportChangeSet(r, branches[2], []string{"br2.txt"}, []string{"br0.txt", "br1.txt"}, t)

	//Modify files
	err = ioutil.WriteFile(path.Join(r.Path, "br0.txt"), []byte("hello world !!"), 0777)
	assertErrNil(err, t)
	_, err = r.addAllAndCommit("commit message")
	assertErrNil(err, t)
	exportChangeSet(r, branches[2], []string{"br2.txt", "br0.txt"}, []string{"br1.txt"}, t)

	//Delete file
	err = os.Remove(path.Join(r.Path, "br1.txt"))
	assertErrNil(err, t)
	_, err = r.addAllAndCommit("commit message")
	exportChangeSet(r, branches[2], []string{"br2.txt", ".wh.br1.txt", "br0.txt"}, []string{"br1.txt"}, t)

	//Uncommited changes
	_, err = r.checkoutB(branches[3])
	assertErrNil(err, t)

	f, err := os.Create(path.Join(r.Path, "br3.txt"))
	assertErrNil(err, t)
	f.Close()
	exportUncommitedChangeSet(r, []string{"br3.txt"}, []string{"br1.txt", ".wh.br1.txt", "br0.txt", "br2.txt"}, t)
	fmt.Printf("OK\n")
}

func exportUncommitedChangeSet(r *gitRepo, expectedFiles, unexpectedFiles []string, t *testing.T) {
	tar, err := r.exportUncommitedChangeSet()
	assertErrNil(err, t)
	defer tar.Close()
	checkTarCorrect(tar, expectedFiles, unexpectedFiles, t)
}

func exportChangeSet(r *gitRepo, branch string, expectedFiles, unexpectedFiles []string, t *testing.T) {
	tar, err := r.exportChangeSet(branch)
	assertErrNil(err, t)
	defer tar.Close()
	checkTarCorrect(tar, expectedFiles, unexpectedFiles, t)
}

func checkTarCorrect(tar archive.Archive, expectedFiles, unexpectedFiles []string, t *testing.T) {
	err := archive.Untar(tar, "/tmp/tar", nil)
	assertErrNil(err, t)
	defer os.RemoveAll("/tmp/tar")
	filesShouldExist(true, expectedFiles, "/tmp/tar", t)
	filesShouldExist(false, unexpectedFiles, "/tmp/tar", t)
}

func filesShouldExist(shouldExist bool, files []string, basePath string, t *testing.T) {
	for _, f := range files {
		exist := fileExists(path.Join(basePath, f))
		if exist != shouldExist {
			t.Fatalf("file %v should exist ? %v", f, shouldExist)
		}
	}
}
