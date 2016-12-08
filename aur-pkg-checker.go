package main

import "fmt"
import "os"
//import "io"
import "io/ioutil"
import "os/exec"
import "bytes"
import "bufio"
import "strings"
import "net/http"
import "regexp"

// execute command and return its sysout
func execute_system_command(command string, arg1 string, arg2 string) string {
  cmd := exec.Command(command, arg1, arg2)
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()
  if err != nil {
    fmt.Fprintln(os.Stderr, "failed to execute command: %v", err)
    os.Exit(1)
  }
  return out.String()
}

func get_installed_pkg_list() map[string]string {
  out := execute_system_command("pacman", "-Q", "-m")
  // Okay, we get the list now. The format is
  // pkg_name pkg_version
  //fmt.Printf("Installed packages: %v\n", out.String())
  scanner := bufio.NewScanner(strings.NewReader(out))
  scanner.Split(bufio.ScanLines)

  // Put installed packages to map
  pkgmap := make(map[string]string)
  for scanner.Scan() {
    results := strings.Split(scanner.Text(), " ")
    pkgname := results[0]
    version := results[1]
    pkgmap[pkgname] = version
    //fmt.Printf("package: %s, version: %s\n", pkgname, version)
  }
  if err := scanner.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading pacman output:", err)
    os.Exit(1)
  }
  return pkgmap
}

func download_pkg_info(pkgname string) map[string]string {
  aur_base_url := "https://aur.archlinux.org/packages/"
  webversion := make(map[string]string)
  response, err := http.Get(aur_base_url + pkgname)
  if err != nil {
    fmt.Println(os.Stderr, "downloading package info:", err)
    os.Exit(1)
  }
  defer response.Body.Close()
  // just convert the html boody to string
  html_body, err := ioutil.ReadAll(response.Body)
  if err != nil {
    fmt.Println(os.Stderr, "parsing response body:", err)
  }
  // use regex to manually parse html
  // this is a bit hacky at the moment
  r, _ := regexp.Compile("<h2>Package Details: (.*)</h2>")
  match := r.Find(html_body)
  versionline := strings.Split(string(match), " ")
  if len(versionline) == 4 {
    //fmt.Println("versionline: " + versionline[3])
    versionstring := strings.Split(versionline[3], "<")
    version := versionstring[0]
    webversion[pkgname] = version
  } else {
    version := "NOT_FOUND"
    webversion[pkgname] = version
  }
  return webversion
}

func get_latest_pkg_versions(installed_map map[string]string) map[string]string {
  latest_versions := make(map[string]string)
  for k, _ := range installed_map {
    //fmt.Println(k,v)
    pkgmap := download_pkg_info(k)
    pkgver := pkgmap[k]
    //fmt.Println("get_latest_pkg_versions: pkgver=" + pkgver)
    latest_versions[k] = pkgver
  }
  return latest_versions
}

func compare_versions(localver map[string]string, webver map[string]string)  {
  for k,_ := range localver {
    //fmt.Println("vercmp", localver[k] + " " + webver[k])
    switch cmp := execute_system_command("vercmp", localver[k], webver[k]); cmp {
    case "-1\n":
        fmt.Println(k + " has new version available (" + webver[k]+ ")")
    case "0\n":
        fmt.Println(k + " version OK")
    case "1\n":
        fmt.Println(k + " maybe OK (" + localver[k] + ")")
    default:
        fmt.Println("Should not reach here! Error, error!")
    }
  }
}

func main()  {
  localpkgs := get_installed_pkg_list()
  //fmt.Println("map: ", pkgs)
  //webversion := download_pkg_info("krop")
  //fmt.Println("webversion: ", webversion)
  webversions := get_latest_pkg_versions(localpkgs)
  compare_versions(localpkgs, webversions)
  //fmt.Println("webversions: ", webversions)
}
