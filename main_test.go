package main

import (
	"strings"
	"testing"
)

func Test_SimilarUniv(t *testing.T) {
	num := computeSimilarUniv("Renmin Univ","Renmin University of China")
	t.Log(num)


	num = computeSimilarUniv("Renmin","Renmin University of China")
	t.Log(num)
}

func Test_GetAbbr(t *testing.T) {
	num := getAbbrName(strings.Split("Renmin University of China"," "))
	t.Log(num)


}