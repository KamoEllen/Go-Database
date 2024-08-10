package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the database at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing collection - no place to save record!")
	}

	if resource == "" {
		return fmt.Errorf("Missing resource - unable to save record (no name)!")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, fnlPath); err != nil {
		return err
	}

	d.log.Info("Successfully wrote data to '%s'\n", fnlPath)
	return nil
}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing collection - unable to read!")
	}

	if resource == "" {
		return fmt.Errorf("Missing resource - unable to read record (no name)!")
	}

	record := filepath.Join(d.dir, collection, resource+".json")

	if _, err := os.Stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (d *Driver) ReadAll(collection string) ([]User, error) {
	if collection == "" {
		return nil, fmt.Errorf("Missing collection - unable to read")
	}
	dir := filepath.Join(d.dir, collection)

	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var users []User

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		var user User
		if err := json.Unmarshal(b, &user); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (d *Driver) Delete(collection, resource string) error {
	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := os.Stat(dir); {
	case err != nil:
		return fmt.Errorf("unable to find file or directory named %v\n", path)
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.Remove(dir + ".json")
	}
	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./Users" // make n add .json

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	employees := []User{
		{"Kamo", "23", "23344333", "RemoteKamo", Address{"Pretoria", "Central", "South Africa", "410013"}},
		{"Kamzo", "25", "23344333", "RemoteKamzo", Address{"Cape Town", "Central", "South Africa", "410013"}},
		{"Kamogelo", "27", "23344333", "RemoteKamogelo", Address{"Durban", "Central", "South Africa", "410013"}},
		{"El", "29", "23344333", "RemoteEL", Address{"Pretoria", "Central", "South Africa", "410013"}},
		{"Ellie", "31", "23344333", "RemoteEllie", Address{"Pretoria", "Central", "South Africa", "410013"}},
		{"Ellen", "32", "23344333", "RemoteEllen", Address{"Pretoria", "Central", "South Africa", "410013"}},
	}

	for _, value := range employees {
		if err := db.Write("users", value.Name, value); err != nil {
			fmt.Println("Error writing user data:", err)
		}
	}

	users, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error reading user data:", err)
		return
	}
	fmt.Println("Records:", users)

	fmt.Println("All Users:")
	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}

	// Optionally delete a user
	if err := db.Delete("users", "John"); err != nil {
		fmt.Println("Error deleting user data:", err)
	}
}
