/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package provisioner

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// WhiteList contains the xname regexp white list.
var WhiteList = make([]string, 0)

// LoadWhiteList loads the white list from a file.
func LoadWhiteList(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Printf("Saved whitelist not found. This is expected if its the first time this has been run.")
		return nil
	}

	f, err := os.Open(filepath.Clean(file))
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Printf("Failed to close whitelist file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		str := scanner.Text()
		WhiteList = append(WhiteList, str)
		log.Printf("DEBUG: Loading %v into Whitelist", str)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Failed to Read Whitelist: %v", err)
		return err
	}

	log.Printf("Loaded Whitelist: %v", WhiteList)

	return nil
}

// AddWhiteListItem adds a xname regexp to the white list.
func AddWhiteListItem(f string, str string) error {
	for _, v := range WhiteList {
		if v == str {
			log.Printf("xname %v already in white list.", str)
			return fmt.Errorf("xname %v already in white list", str)
		}
	}

	WhiteList = append(WhiteList, str)

	if err := WriteWhiteList(f); err != nil {
		return err
	}

	log.Printf("xname %v added to white list.", str)

	return nil
}

// RemoveWhiteListItem removes a xname regexp from the white list.
func RemoveWhiteListItem(f string, str string) error {
	exists := false

	for i, v := range WhiteList {
		if v == str {
			WhiteList[i] = WhiteList[len(WhiteList)-1]
			WhiteList[len(WhiteList)-1] = ""
			WhiteList = WhiteList[:len(WhiteList)-1]
			exists = true
		}
	}

	if !exists {
		log.Printf("xname %v is not in white list", str)
		return fmt.Errorf("xname %v is not in white list", str)
	}

	if err := WriteWhiteList(f); err != nil {
		return err
	}

	log.Printf("xname %v removed from whitelist", str)

	return nil
}

// WriteWhiteList saves the white list to a file.
func WriteWhiteList(file string) error {
	f, err := os.Create(filepath.Clean(file))
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Printf("Failed to close whitelist file: %v", err)
		}
	}()

	writer := bufio.NewWriter(f)

	log.Printf("Writing WhiteList: %v", WhiteList)

	for _, v := range WhiteList {
		log.Printf("Debug: Write to %v: %v", file, v)

		_, err := writer.WriteString(v + "\n")
		if err != nil {
			return err
		}

		err = writer.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// validateXname validates that an xname matches a regexp in the white list.
func validateXname(xname string) error {
	exists := false

	for _, v := range WhiteList {
		r, err := regexp.Compile(v)
		if err != nil {
			return err
		}

		if r.MatchString(xname) {
			exists = true
		}
	}

	if !exists {
		log.Printf("xname %v is not white listed", xname)
		return fmt.Errorf("xname %v is not white listed", xname)
	}

	return nil
}
