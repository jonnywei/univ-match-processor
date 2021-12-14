package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	currDir, _ := os.Getwd()
	src := flag.String("src", currDir, "csv file dir")
	flag.Parse()
	currDir = *src
	files, err := ioutil.ReadDir(currDir)
	if err != nil {
		log.Fatal(err)
	}

	univPath := currDir + "/" + "机构排名.csv"
	genPath := currDir + "/new_dir"
	os.MkdirAll(genPath, os.ModeDir)
	totalUnviList := buildUniversityList(univPath)
	for _, f := range files {
		var fileName = f.Name()
		if !strings.HasSuffix(fileName, ".csv") || strings.EqualFold(fileName, "机构排名.csv") {
			continue
		}

		fmt.Println("process file " + fileName)
		srcPath := currDir + "/" + fileName
		dstPath := genPath + "/机构排名-" + fileName
		univList := buildUniversityList(univPath)
		processOneFile(&totalUnviList, univList, srcPath, dstPath)
	}
	writeUniversityList(&totalUnviList, univPath)
	fmt.Println("Press any key to close")
	bufio.NewReader(os.Stdin).ReadRune()

}

type University struct {
	name        string
	aliasName   []string
	country     string
	similarName map[string]string
	count       int
}

type Country struct {
	name      string
	aliasName []string
}

func buildUniversityList(srcPath string) []University {
	var univList []University
	records, _ := readCsvFile(srcPath)
	for index := range records {
		if index == 0 {
			continue
		}
		row := records[index]
		name := records[index][0]
		u := University{
			name:        name,
			count:       0,
			similarName: make(map[string]string),
		}
		if len(row) >= 2 {
			u.country = records[index][1]
		}
		if len(row) >= 3 {
			u.aliasName = strings.Split(records[index][2], ";")
		}

		univList = append(univList, u)
	}
	return univList
}

func writeUniversityList(univList *[]University, dstPath string) {
	var results = make([][]string, len(*univList)+1)
	row := []string{"UnivFullName" + time.Now().Format("15:04:05"), "Country", "UnivAliasName", "UnivSimilarName", "count"}
	results[0] = row
	for index, u := range *univList {
		row := []string{
			u.name,
			u.country,
			strings.Join(u.aliasName, ";"),
			printMapKey(u.similarName),
			//fmt.Sprintf("%v",u.similarName),
			fmt.Sprintf("%d", u.count)}
		results[index+1] = row
	}
	writeCsvFile(dstPath, results)
}

func processOneFile(univList *[]University, currentUnivList []University, srcPath string, dstPath string) {

	records, _ := readCsvFile(srcPath)
	coutinue_empty := 0
	////row = append(row, "Univ_Sem", "Univ_Sem", "Univ_Sem")
	//row = append(row, FormatBool(true), FormatBool(true), FormatBool(true))
	//results[index] = row
	for index := range records {
		if index == 0 {
			continue
		}
		if len(records[index][0]) == 0 {
			coutinue_empty += 1
			fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!EMPTY LINE!!!!!!!!!!!!!!!!!!!!!")
			if coutinue_empty > 10 {
				fmt.Printf("line %d: empty !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!EMPTY LINE!!!!!!!!!!!!!!!!!!!!! break\n", index)
				break
			}
			continue
		}
		coutinue_empty = 0
		addresses := processRecord(records[index][0])
		fmt.Printf("Record:%d ==============START=================\n", index)
		firstAuthorAddr := getFirstAuthorAddr(addresses)
		fmt.Println("firstAuthorAddr:" + firstAuthorAddr)
		univ, country := getUnivAndCountry(firstAuthorAddr)
		fmt.Printf("UV:{%s}:{%s}\n", univ, country)
		findUniv(univ, country, univList)
		findUniv(univ, country, &currentUnivList)
	}
	writeUniversityList(&currentUnivList, dstPath)

}

func readCsvFile(filePath string) ([][]string, error) {
	f, error := os.Open(filePath)
	if error != nil {
		fmt.Println("unable to open file:", filePath)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)

	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println("unable to open csv reader:", err)
	}
	return records, nil
}

func writeCsvFile(filePath string, records [][]string) ([][]string, error) {
	f, error := os.Create(filePath)
	if error != nil {
		fmt.Printf("unable to writeCsvFile:{%s}\n", filePath)
	}
	defer f.Close()
	csvWriter := csv.NewWriter(f)

	err := csvWriter.WriteAll(records)
	if err != nil {
		fmt.Printf("ERROR: unable to open csv writeCsvFile:{%s}\n", err)
	}
	return records, nil
}

//record formats
/**
[Way, Sean A.; Tracey, J. Bruce] Cornell Univ, Ithaca, NY 14853 USA;
[Fay, Charles H.] Rutgers State Univ, Piscataway, NJ 08855 USA;
[Wright, Patrick M.] Univ S Carolina, Columbia, SC 29208 USA;
[Snell, Scott A.] Univ Virginia, Charlottesville, VA 22903 USA;
[Chang, Song] Chinese Univ Hong Kong, Hong Kong, Hong Kong, Peoples R China;
[Gong, Yaping] Hong Kong Univ Sci & Technol, Hong Kong, Hong Kong, Peoples R China
*/

func processRecord(address string) []string {
	reg := regexp.MustCompile("\\[[^\\]]*\\]")
	remain := reg.ReplaceAllString(address, "")
	//fmt.Println(remain)
	addressList := strings.Split(remain, ";")
	return addressList
}
func getFirstCorrespondingAuthorAddr(address string) string {
	reg := regexp.MustCompile("\\[[^\\]]*\\]")
	remain := reg.ReplaceAllString(address, "")
	addressList := strings.Split(remain, ";")
	return addressList[0]
}

func getUnivAndCountry(address string) (string, string) {
	addr := strings.Split(address, ",")
	length := len(addr)
	if length < 2 {
		fmt.Errorf("address error: %s", address)
		return "", ""
	}
	univ := addr[0]
	country := addr[length-1]
	return strings.TrimSpace(univ), strings.TrimSpace(country)
}

//Texas Christian Univ
func findUniv(univ string, country string, univList *[]University) bool {

	for index, _ := range *univList {
		var isFind = false
		u := &(*univList)[index]
		if strings.EqualFold(univ, u.name) {
			isFind = true
		}
		for _, aliasName := range u.aliasName {
			aliasName = strings.TrimSpace(aliasName)
			if strings.EqualFold(univ, aliasName) {
				isFind = true
				if isFind {
					fmt.Println("----------=============================------------")
				}
				break
			}
		}
		if compareCountry(u.country, country) {
			if isFind {
				u.count += 1
				return true
			}
			if computeSimilarUniv(univ, u.name) {
				u.similarName[univ] = univ
			}
		} else {
			fmt.Println(u.country + " not equal " + country)
		}
	}

	return false
}

//Texas Christian Univ
func computeSimilarUniv(univ string, fullName string) bool {
	newFullName := strings.ReplaceAll(fullName, "The", "")
	newFullName = strings.ReplaceAll(newFullName, " of", "")
	newFullName = strings.ReplaceAll(newFullName, "University", "Univ")
	newFullName = strings.ReplaceAll(newFullName, " at ", " ")
	//newFullName  = strings.ReplaceAll(newFullName,"at","Univ")
	//newFullName  = strings.ReplaceAll(newFullName,"University","Univ")
	fmt.Println(univ + " Compare: " + newFullName)
	fullNameArray := strings.Split(newFullName, " ")
	univArray := strings.Split(univ, " ")
	if len(univArray) == 1 {
		if univArray[0] == getAbbrName(fullNameArray) {
			return true
		}
	}
	similarCount := 0
	fullNameCount := len(fullNameArray)
	for _, u := range univArray {
		if stringInArray(u, fullNameArray) {
			if u == "Univ" {
				fullNameCount -= 1
			} else {
				similarCount += 1
			}
		}
	}
	simV := float32(similarCount) / float32(fullNameCount)
	fmt.Println(similarCount, fullNameCount)
	if simV > 0.5 {
		return true
	}
	return false
}

func getAbbrName(fullNameArray []string) string {
	var abbr string
	for _, t := range fullNameArray {
		t := strings.TrimSpace(t)
		if t != "" {
			abbr = abbr + t[0:1]
		}
	}
	return abbr
}

func compareCountry(c string, country string) bool {
	china := Country{name: "China", aliasName: []string{"Peoples R China", "China", "中国"}}
	uk := Country{name: "UK", aliasName: []string{"UK", "United Kingdom", "England", "英国"}}
	countriesList := map[string]Country{"china": china, "UK": uk}
	//fmt.Println(countriesList)
	if len(c) == 0 || len(strings.TrimSpace(c)) == 0 {
		return true
	}
	if strings.Contains(country, c) {
		return true
	}
	if val, ok := countriesList[c]; ok {
		//do something here
		if stringInArray(country, val.aliasName) {
			return true
		}
	}
	return false

}

func stringInArray(word string, array []string) bool {
	if len(array) == 0 {
		return false
	}
	for _, w := range array {
		if strings.EqualFold(word, w) {
			return true
		} else {
			if len(word) > 2 && strings.HasSuffix(w, word) {
				return true
			}
		}
	}
	return false
}

//Chinese Univ Hong Kong, Sch Management & Econ, Shenzhen 518172, Guangdong, Peoples R China;
//Univ Macau, Fac Educ, Room J507A,5th Floor,Silver Jubilee Bldg, Taipa, Macau, Peoples R China.;
//Meyer, JP (corresponding author), Univ Western Ontario, Dept Psychol, Social Sci Ctr, London, ON N6A 5C2, Canada.
//[Lam, Long Wai] Univ Macau, Taipa, Macao, Peoples R China;
//[Peng, Kelly Z.] Hong Kong Shue Yan Univ, North Point, Hong Kong, Peoples R China;
//[Wong, Chi-Sum; Lau, Dora C.] Chinese Univ Hong Kong, Hong Kong, Hong Kong, Peoples R China

func getFirstAuthorAddr(addressList []string) string {
	if len(addressList) == 0 {
		return ""
	}
	return strings.TrimSpace(addressList[0])
}

func FormatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func printMapKey(sm map[string]string) string {
	keys := make([]string, 0)
	var str string = ""
	for k, _ := range sm {

		str = str + k + ";"
		keys = append(keys, k)
	}
	//fmt.Printf("keys:%d\n", len(keys))
	return strings.Join(keys, ";")
}
