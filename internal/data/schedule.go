package data

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Days of the week
var weekDays = [7]string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}

// type strings
type Day []string

type IntervalPeriod struct {
	IntervalID int    `json:"interval_id"`
	Interval   string `json:"interval"`
}
type ScheduleModel struct {
	DB *sql.DB
}

type Days struct {
	Day       string      `json:"day"`
	Intervals []Intervals `json:"intervals"`
}
type Intervals struct {
	IntervalID  int    `json:"interval_id"`
	CourseID    int    `json:"course_id"`
	Description string `json:"description"`
}

// Struct to read schedule data from user
type Schedule struct {
	ProgramID  int    `json:"program_id"`
	SemesterID int    `json:"semester_id"`
	Days       []Days `json:"days"`
}

type StudentInterval struct {
	Interval    string `json:"interval"`
	CourseID    int    `json:"course_id"`
	CourseCode  string `json:"course_code"`
	CourseTitle string `json:"course_title"`
	TeacherID   int    `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	Description string `json:"description"`
}

type TeacherInterval struct {
	Interval    string `json:"interval"`
	CourseID    int    `json:"course_id"`
	CourseCode  string `json:"course_code"`
	CourseTitle string `json:"course_title"`
	Description string `json:"description"`
}

type StudentDay struct {
	Day       string            `json:"day"`
	Intervals []StudentInterval `json:"intervals"`
}
type StudentSchedule struct {
	Days []StudentDay `json:"days"`
}

type TeacherDay struct {
	Day       string            `json:"day"`
	Intervals []TeacherInterval `json:"intervals"`
}

type TeacherSchedule struct {
	Days []TeacherDay `json:"days"`
}

// GetAllDays returns a list of days
func (m ScheduleModel) GetAllDays() (*Day, error) {

	// Construct query
	query := `SELECT day FROM days`
	var d Day

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	defer rows.Close()

	// loop through rows
	for rows.Next() {
		var temp string
		err := rows.Scan(&temp)
		if err != nil {
			return nil, err
		}

		// Add to days
		d = append(d, temp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return the days
	return &d, nil
}

// GetAllIntervals returns a list of intervals
func (m ScheduleModel) GetAllIntervals() (*[]IntervalPeriod, error) {

	// Construct query
	query := `SELECT interval_id, interval FROM intervals`

	var periods []IntervalPeriod

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	defer rows.Close()

	// loop through rows
	for rows.Next() {
		var temp IntervalPeriod
		err := rows.Scan(&temp.IntervalID, &temp.Interval)

		if err != nil {
			return nil, err
		}

		// Add to periods
		periods = append(periods, temp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return the days
	return &periods, nil
}

// SetSchedule sets a schedule for a semester of a program
func (m ScheduleModel) SetSchedule(obj *Schedule) error {

	programID := obj.ProgramID
	semesterID := obj.SemesterID

	query1 := `SELECT program_id, semester_id FROM running_semesters 
	WHERE program_id = $1 AND semester_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var prog, sem int
	row := m.DB.QueryRowContext(ctx, query1, programID, semesterID)
	// Scan the values
	row.Scan(&prog, &sem)

	// If values stay zero, that means no such mix of prog and sem ids exists
	if prog == 0 && sem == 0 {
		return ErrNoRecords
	}
	// program_id | semester_id |  day   | interval_id | course_id | description
	query := `INSERT INTO day_schedule (program_id, semester_id, day, interval_id, course_id, description) VALUES `

	args := []interface{}{}

	for _, val1 := range obj.Days {

		day := val1.Day
		intervals := val1.Intervals

		for _, val2 := range intervals {

			args = append(args, programID, semesterID, day, val2.IntervalID, val2.CourseID, val2.Description)
		}
	}

	// No of values to be inserted
	length := len(args) / 6
	var values []string

	// Since number of values is not known, we will construct query dynamically
	for i := 0; i < length; i++ {
		temp := ` (`
		for j := i*6 + 1; j < ((i+1)*6)+1; j++ {
			if j == (i+1)*6 {
				temp += `$` + strconv.Itoa(j)
				continue
			}
			temp += `$` + strconv.Itoa(j) + `, `
		}
		temp += `)`
		values = append(values, temp)
	}

	otherQuery := strings.Join(values, ",")

	// The whole query
	fullQuery := query + otherQuery

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, fullQuery, args...)

	if err != nil {
		// Incase of duplicate entry
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "day_schedule_pkey"`:
			return ErrDuplicateEntry
		default:
			return err
		}
	}
	// Success
	return nil
}

func (m ScheduleModel) GetSchedule(programID, semesterID int) (*StudentSchedule, error) {

	var obj StudentSchedule

	query1 := `SELECT program_id, semester_id FROM running_semesters 
	WHERE program_id = $1 AND semester_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var prog, sem int
	row := m.DB.QueryRowContext(ctx, query1, programID, semesterID)
	// Scan the values
	row.Scan(&prog, &sem)

	// If values stay zero, that means no such mix of prog and sem ids exists
	if prog == 0 && sem == 0 {
		return nil, ErrNoRecords
	}

	query := `SELECT day_schedule.program_id, day_schedule.semester_id,
	day_schedule.day, day_schedule.description, intervals.interval,
	courses.course_id, courses.course_code,
	courses.title ,teachers.teacher_id, teachers.name
	FROM day_schedule
	INNER JOIN courses on courses.course_id = day_schedule.course_id
	INNER JOIN intervals on intervals.interval_id = day_schedule.interval_id
	INNER JOIN teacher_courses on teacher_courses.course_id = day_schedule.course_id
	INNER JOIN teachers on teachers.teacher_id = teacher_courses.teacher_id
	WHERE day_schedule.program_id = $1 AND day_schedule.semester_id = $2`

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx1, query, programID, semesterID)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}
	defer rows.Close()

	allSchedule := map[string][]StudentInterval{}
	// loop through rows
	for rows.Next() {
		//  program_id | semester_id |  day   |    description     |  interval   | course_id | course_code |title| teacher_id |     name
		var day string
		var temp StudentInterval
		var programID, semesterID int

		err := rows.Scan(&programID, &semesterID, &day, &temp.Description, &temp.Interval, &temp.CourseID,
			&temp.CourseCode, &temp.CourseTitle, &temp.TeacherID, &temp.TeacherName)

		if err != nil {
			return nil, err
		}

		allSchedule[day] = append(allSchedule[day], temp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, val := range weekDays {
		var temp StudentDay

		data, exists := allSchedule[val]
		// If a certain day has no entries
		if !exists {
			temp.Day = val
			temp.Intervals = nil

			obj.Days = append(obj.Days, temp)
			continue
		}

		// if exists
		temp.Day = val

		for _, x := range data {

			temp.Intervals = append(temp.Intervals, x)
		}

		obj.Days = append(obj.Days, temp)
	}

	return &obj, nil
}

// GetTeacherSchedule returns the schedule for a teacher
func (m ScheduleModel) GetTeacherSchedule(userID int) (*TeacherSchedule, error) {

	var obj TeacherSchedule
	query := `
	SELECT day_schedule.program_id, day_schedule.semester_id,
	day_schedule.day, day_schedule.description, intervals.interval,
	courses.course_id, courses.course_code,
	courses.title FROM day_schedule
	INNER JOIN courses on courses.course_id = day_schedule.course_id
	INNER JOIN intervals on intervals.interval_id = day_schedule.interval_id
	INNER JOIN teacher_courses on teacher_courses.course_id = day_schedule.course_id
	INNER JOIN teachers on teachers.teacher_id = teacher_courses.teacher_id
	WHERE teachers.user_id = $1`

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx1, query, userID)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}
	defer rows.Close()

	allSchedule := map[string][]TeacherInterval{}
	// loop through rows
	for rows.Next() {
		//  program_id | semester_id |  day   |    description     |  interval   | course_id | course_code |title
		var day string
		var temp TeacherInterval
		var programID, semesterID int
		var desc sql.NullString // Used for column that might return null or string
		err := rows.Scan(&programID, &semesterID, &day, &desc, &temp.Interval, &temp.CourseID,
			&temp.CourseCode, &temp.CourseTitle)

		if err != nil {
			return nil, err
		}
		if desc.Valid { //if not null use this
			temp.Description = desc.String
		} else {
			temp.Description = ""
		}

		allSchedule[day] = append(allSchedule[day], temp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, val := range weekDays {
		var temp TeacherDay

		data, exists := allSchedule[val]
		// If a certain day has no entries
		if !exists {
			temp.Day = val
			temp.Intervals = nil

			obj.Days = append(obj.Days, temp)
			continue
		}

		// if exists
		temp.Day = val

		for _, x := range data {

			temp.Intervals = append(temp.Intervals, x)
		}

		obj.Days = append(obj.Days, temp)
	}

	return &obj, nil
}

// GetStudentSchedule returns the schedule for a student
func (m ScheduleModel) GetStudentSchedule(userID int) (*StudentSchedule, error) {

	var obj StudentSchedule

	query := `SELECT day_schedule.program_id, day_schedule.semester_id,
	day_schedule.day, day_schedule.description, intervals.interval,
	courses.course_id, courses.course_code,
	courses.title ,teachers.teacher_id, teachers.name
	FROM day_schedule
	INNER JOIN courses on courses.course_id = day_schedule.course_id
	INNER JOIN intervals on intervals.interval_id = day_schedule.interval_id
	INNER JOIN teacher_courses on teacher_courses.course_id = day_schedule.course_id
	INNER JOIN teachers on teachers.teacher_id = teacher_courses.teacher_id
	WHERE day_schedule.program_id = (SELECT students.program_id FROM students WHERE students.user_id = $1)
	 AND day_schedule.semester_id = (SELECT students.semester_id FROM students WHERE students.user_id = $1)`

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx1, query, userID)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}
	defer rows.Close()

	allSchedule := map[string][]StudentInterval{}
	// loop through rows
	for rows.Next() {
		//  program_id | semester_id |  day   |    description     |  interval   | course_id | course_code |title| teacher_id |     name
		var day string
		var temp StudentInterval
		var programID, semesterID int
		var desc sql.NullString // Used for column that might return null or string
		err := rows.Scan(&programID, &semesterID, &day, &desc, &temp.Interval, &temp.CourseID,
			&temp.CourseCode, &temp.CourseTitle, &temp.TeacherID, &temp.TeacherName)

		if err != nil {
			return nil, err
		}
		if desc.Valid { //if not null use this
			temp.Description = desc.String
		} else {
			temp.Description = ""
		}

		allSchedule[day] = append(allSchedule[day], temp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, val := range weekDays {
		var temp StudentDay

		data, exists := allSchedule[val]
		// If a certain day has no entries
		if !exists {
			temp.Day = val
			temp.Intervals = nil

			obj.Days = append(obj.Days, temp)
			continue
		}

		// if exists
		temp.Day = val

		for _, x := range data {

			temp.Intervals = append(temp.Intervals, x)
		}

		obj.Days = append(obj.Days, temp)
	}

	return &obj, nil
}

// DeleteSchedule deletes a schedule for a semester of a program
func (m ScheduleModel) DeleteSchedule(programID, semesterID int) error {

	query1 := `DELETE FROM day_schedule WHERE program_id= $1 AND semester_id= $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := m.DB.ExecContext(ctx, query1, programID, semesterID)

	if err != nil {
		return err
	}

	affected, err := row.RowsAffected()

	if err != nil {
		return err
	}

	// If affected rows = 0 then the notice did not exist
	if affected == 0 {
		return ErrRecordNotFound
	}

	// Success
	return nil
}
