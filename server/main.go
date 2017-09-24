package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type InputSet struct {
	Name     string `yaml:"name"`
	Problem  string `yaml:"problem"`
	Solution string `yaml:"solution"`
}

type InputSets []InputSet

func (sets InputSets) Get(name string) (set InputSet, exists bool) {
	for _, s := range sets {
		if s.Name == name {
			set = s
			exists = true
		}
	}
	return
}

type StudentOutputs struct {
	Submitted int
	Correct   int
}

type Student struct {
	Outputs map[string]StudentOutputs
	Team    string
}
type StudentMap map[string]Student

func (smap StudentMap) GetStudent(name string) Student {
	_, exists := smap[name]
	if !exists {
		student := Student{}
		student.Outputs = map[string]StudentOutputs{}
		smap[name] = student
	}
	return smap[name]
}

func main() {
	teams := []string{"Orange", "Green"}

	var sets InputSets
	var data []byte
	students := StudentMap{}

	{
		var err error
		data, err = ioutil.ReadFile("./inputsets.yml")
		if err != nil {
			logrus.Fatal(err)
		}
	}
	{
		err := yaml.Unmarshal(data, &sets)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/uuid", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(uuid.NewV4().String()))
	})
	r.GET("/recv/:uuid/:name", func(c *gin.Context) {
		name := c.Param("name")
		uuid := c.Param("uuid")
		logrus.Info(uuid)

		// Special case: request team
		if name == "team" {
			student := students.GetStudent(uuid)
			if student.Team == "" {
				student.Team = teams[rng.Int()%len(teams)]
				students[uuid] = student
			}
			c.String(http.StatusOK, "You are on team "+student.Team+"\n")
			return
		}

		set, exists := sets.Get(name)
		if !exists {
			c.String(http.StatusBadRequest, "Unrecognized input set: '"+name+"'")
			return
		}
		c.String(http.StatusOK, set.Problem)
	})
	r.POST("/send/:uuid/:name", func(c *gin.Context) {
		name := c.Param("name")
		uuid := c.Param("uuid")
		logrus.Info(uuid)
		student := students.GetStudent(uuid)
		if _, exists := student.Outputs[name]; !exists {
			student.Outputs[name] = StudentOutputs{}
		}
		outputs := student.Outputs[name]
		outputs.Submitted++

		student.Outputs[name] = outputs
		students[uuid] = student

		output := c.PostForm("data")
		set, exists := sets.Get(name)
		if !exists {
			c.String(http.StatusBadRequest, "Unrecognized input set: '"+name+"'")
			return
		}

		userString := strings.Trim(output, " \n")
		soluString := strings.Trim(set.Solution, " \n")

		if userString == soluString {
			outputs.Correct++

			student.Outputs[name] = outputs
			students[uuid] = student
			c.String(http.StatusOK, "Success!\n")
		} else {
			c.String(http.StatusOK, "Nope! Try again :P\n")
		}
	})
	r.GET("/display/:name", func(c *gin.Context) {
		var studentsSubmitted int
		var studentsCorrect int
		var totalSubmitted int
		var totalCorrect int
		teamScores := map[string]int{}

		name := c.Param("name")

		for _, team := range teams {
			teamScores[team] = 0
		}

		for _, student := range students {
			outputs, exists := student.Outputs[name]
			if !exists {
				outputs = StudentOutputs{}
			}

			if outputs.Submitted > 0 {
				studentsSubmitted++
			}
			if outputs.Correct > 0 {
				studentsCorrect++
				teamScores[student.Team]++
			}
			totalSubmitted += outputs.Submitted
			totalCorrect += outputs.Correct
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"students_submitted": studentsSubmitted,
			"students_correct":   studentsCorrect,
			"total_submitted":    totalSubmitted,
			"total_correct":      totalCorrect,
			"teams":              teamScores,
		})
	})
	r.Run()
}
