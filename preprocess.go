package carstore

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cbergoon/merkletree"
)

func JsonToByte(filepath string) string {
	//打开文件
	inputFile, inputError := os.Open(filepath)
	if inputError != nil {
		return "error"
	}
	defer inputFile.Close()
	//按行读取json文件
	var s string
	inputReader := bufio.NewReader(inputFile)
	for {
		inputString, readerError := inputReader.ReadString('\n')
		if readerError == io.EOF {
			break
		}
		s = s + inputString
	}
	//去空格
	s = strings.Replace(s, "\r\n", "", -1)
	s = strings.Replace(s, " ", "", -1)
	//str := str2bytes(s)
	fmt.Printf("%s", s)
	return s
}

//--------------------------------------------------------------------------------//
//以上代码完成字符串转换
//TestContent implements the Content interface provided by merkletree and represents the content stored in the tree.
type TestContent struct {
	x string
}

//merkle树给出的两个接口实现
//CalculateHash hashes the values of a TestContent
func (t TestContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

//Equals tests for equality of two Contents
func (t TestContent) Equals(other merkletree.Content) (bool, error) {
	return t.x == other.(TestContent).x, nil
}

func GenMerkleRoot() []byte {
	//Build list of Content to build tree
	//用字符串构造默克尔树
	var list []merkletree.Content
	var str string
	str = JsonToByte()
	list = append(list, TestContent{x: str})
	list = append(list, TestContent{x: str})
	list = append(list, TestContent{x: str})
	list = append(list, TestContent{x: "123"})

	//Create a new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		//log.Fatal(err)
	}

	//Get the Merkle Root of the tree
	mr := t.MerkleRoot()
	//fmt.Println()
	//fmt.Println(mr)
	return mr
}
