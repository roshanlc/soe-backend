package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// A struct to hold offered by details for a course
type OfferedBy struct {
	Faculty    string `json:"faculty"`
	Department string `json:"department"`
	Program    string `json:"program"`
	Level      string `json:"level"`
	Semester   int    `json:"semester"`
}

// A struct to hold information about book
type Book struct {
	BookID      int64  `json:"book_id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Edition     int    `json:"edition"`
	Publication string `json:"publication"`
}

// A struct to hold info about teachers teaching a course
type TaughtBy struct {
	TeacherID int64  `json:"teacher_id"`
	Name      string `json:"name"`
}

// A struct to hold information about courses
type Course struct {
	CourseID   int64       `json:"course_id"`
	CourseCode string      `json:"course_code"`
	Title      string      `json:"title"`
	Credit     int         `json:"credit"`
	Elective   bool        `json:"elective"`
	OfferedBy  []OfferedBy `json:"offered_by"` // Which program at which sem offeres this course
	TextBooks  []Book      `json:"text_books"` // The Text books for this course
	RefBooks   []Book      `json:"ref_books"`  // Reference Books for this course
	TaughtBy   []TaughtBy  `json:"taught_by"`  // Who teaches this course
}

// Wrapper for Course having a *sql.DB connection
type CourseModel struct {
	DB *sql.DB
}

// Returns a specific course
// else an error incase if any
func (m CourseModel) Get(courseCode string) (*Course, error) {

	// Construct query to get the info
	/*
		query := `SELECT courses.course_id, courses.course_code, courses.title, courses.credit, courses.elective,
		faculties.name as faculty_name, departments.name as dept_name, programs.name as program, levels.name as level,
		semesters.semester_id as semester, books.book_id,books.title as book_title, books.author, books.edition, books.publication,
		course_books.text_book, teachers.teacher_id, teachers.name as teacher_name
		FROM courses
		INNER JOIN program_courses ON program_courses.course_id = courses.course_id
		INNER JOIN programs ON programs.program_id = program_courses.program_id
		INNER JOIN departments ON departments.department_id =  programs.department_id
		INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id
		INNER JOIN levels ON levels.level_id = programs.program_id
		INNER JOIN running_semesters ON running_semesters.program_id = programs.program_id
		INNER JOIN  semesters ON semesters.semester_id = running_semesters.semester_id
		INNER JOIN  course_books ON course_books.course_id = courses.course_id
		INNER JOIN books ON books.book_id = course_books.book_id
		INNER JOIN  teacher_courses ON teacher_courses.course_id = courses.course_id
		INNER JOIN  teachers  ON teachers.teacher_id = teachers.teacher_id
		WHERE LOWER(courses.course_code) = LOWER($1)`
	*/

	// Construct multiple queries instead of one
	// Multiple queries almost 3x fast the single query
	// Since connection time to db is negligible in our case

	// For courses records
	courseQuery := `SELECT courses.course_id, courses.course_code, courses.title, courses.credit, courses.elective
	 FROM courses WHERE LOWER(courses.course_code) = LOWER($1)`

	// For text books of a course
	bookQuery := `SELECT courses.course_id,books.book_id, books.title, books.author, books.edition, books.publication, course_books.text_book
	FROM books
	INNER JOIN course_books ON course_books.book_id = books.book_id
	INNER JOIN courses ON courses.course_id = course_books.course_id
	WHERE LOWER(courses.course_code) = LOWER($1)`

	// Who teaches that course

	teacherQuery := `SELECT courses.course_id,teachers.teacher_id, teachers.name as teacher
	FROM teachers
	INNER JOIN teacher_courses ON teachers.teacher_id = teacher_courses.teacher_id
	INNER JOIN courses ON courses.course_id = teacher_courses.course_id
	WHERE LOWER(courses.course_code) = LOWER($1)`

	// which faculty offeres this course

	offeredByQuery := `SELECT courses.course_id, faculties.name as faculty, departments.name as dept, programs.name as program, levels.name as level,
	program_courses.semester_id as semester
	FROM courses
	INNER JOIN program_courses ON courses.course_id = program_courses.course_id
	INNER JOIN programs ON programs.program_id = program_courses.program_id
	INNER JOIN levels ON levels.level_id = programs.level_id
	INNER JOIN departments ON departments.department_id = programs.department_id
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id
	WHERE LOWER(courses.course_code) = LOWER($1)`

	// Create a time out context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Defer cancel of transaction
	defer cancel()

	// Empty course
	var course Course

	// Query the db for course's detail
	err := m.DB.QueryRowContext(ctx, courseQuery, courseCode).Scan(
		&course.CourseID,
		&course.CourseCode,
		&course.Title,
		&course.Credit,
		&course.Elective)

	// if any error has occured
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Now that course exists on db, we query for more details

	bookRows, err := m.DB.QueryContext(ctx, bookQuery, courseCode)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any books listed
		default:
			return nil, err

		}
	}

	// Close the bookRows after reading
	defer bookRows.Close()

	// Loop through rows
	for bookRows.Next() {

		// temp book  struct
		var book Book

		// course id to maintain reference
		var courseID int64
		// boolean to check if a book is text book
		var isTextBook bool

		// SELECT courses.course_id,books.book_id, books.title, books.author,
		// books.edition, books.publication, course_books.text_book
		err := bookRows.Scan(&courseID,
			&book.BookID,
			&book.Title,
			&book.Author,
			&book.Edition,
			&book.Publication,
			&isTextBook)

		// If some error occurred
		if err != nil {
			return nil, err
		}

		if isTextBook {
			course.TextBooks = append(course.TextBooks, book)
		} else {
			course.RefBooks = append(course.RefBooks, book)
		}

	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = bookRows.Err(); err != nil {
		return nil, err
	}

	// Query for teachers
	teacherRows, err := m.DB.QueryContext(ctx, teacherQuery, courseCode)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any teachers listed
		default:
			return nil, err

		}
	}

	// Defer closing the teacherRows
	defer teacherRows.Close()

	// Loop through teacher rows
	for teacherRows.Next() {

		// temp TaughtBy  struct
		var teacher TaughtBy

		// course id to maintain reference
		var courseID int64

		// SELECT courses.course_id,teachers.teacher_id, teachers.name as teacher
		err := teacherRows.Scan(
			&courseID,
			&teacher.TeacherID,
			&teacher.Name)
		// Incase any err occurred
		if err != nil {
			return nil, err
		}

		// Add the teacher to the course object
		course.TaughtBy = append(course.TaughtBy, teacher)
	}

	// Incase any error occurred while going through the loop
	if err = teacherRows.Err(); err != nil {
		return nil, err
	}

	// Query the db to get who offeres the subject

	// Query for teachers
	offeredByRows, err := m.DB.QueryContext(ctx, offeredByQuery, courseCode)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any offered by listed
		default:
			return nil, err

		}
	}

	// Defer closing the teacherRows
	defer offeredByRows.Close()

	// Loop through teacher rows
	for offeredByRows.Next() {

		// temp offeredBy  struct
		var offeredBy OfferedBy

		// course id to maintain reference
		var courseID int64

		// SELECT courses.course_id, faculties.name as faculty,
		// departments.name as dept, programs.name as program, levels.name as level, program_courses.semester_id as semester
		err := offeredByRows.Scan(
			&courseID,
			&offeredBy.Faculty,
			&offeredBy.Department,
			&offeredBy.Program,
			&offeredBy.Level,
			&offeredBy.Semester)

		// Incase any err occurred
		if err != nil {
			return nil, err
		}

		// Add the teacher to the course object
		course.OfferedBy = append(course.OfferedBy, offeredBy)
	}

	// Incase any error occurred while going through the loop
	if err = offeredByRows.Err(); err != nil {
		return nil, err
	}

	// Finally return the course
	return &course, nil

}

// Gets all courses by the provided filters
func (m CourseModel) GetAll(faculty, department, program, level string, semester int) ([]Course, error) {

	// For now no search paramters will be implemented

	// For courses records
	courseQuery := `SELECT courses.course_id, courses.course_code, courses.title, courses.credit, courses.elective
	 FROM courses`

	// For text books of a course
	bookQuery := `SELECT courses.course_id,books.book_id, books.title, books.author, books.edition, books.publication, course_books.text_book
	FROM books
	INNER JOIN course_books ON course_books.book_id = books.book_id
	INNER JOIN courses ON courses.course_id = course_books.course_id`

	// Who teaches that course

	teacherQuery := `SELECT courses.course_id,teachers.teacher_id, teachers.name as teacher
	FROM teachers
	INNER JOIN teacher_courses ON teachers.teacher_id = teacher_courses.teacher_id
	INNER JOIN courses ON courses.course_id = teacher_courses.course_id`

	// which faculty offeres this course

	offeredByQuery := `SELECT courses.course_id, faculties.name as faculty, departments.name as dept, programs.name as program, levels.name as level,
	program_courses.semester_id as semester
	FROM courses
	INNER JOIN program_courses ON courses.course_id = program_courses.course_id
	INNER JOIN programs ON programs.program_id = program_courses.program_id
	INNER JOIN levels ON levels.level_id = programs.level_id
	INNER JOIN departments ON departments.department_id = programs.department_id
	INNER JOIN faculties ON faculties.faculty_id = departments.faculty_id`

	// Create a time out context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Defer cancel of transaction
	defer cancel()

	// all courses will be mapped here, and data will be updated through course_id reference
	var allCourses = make(map[int64]Course)

	courseRows, err := m.DB.QueryContext(ctx, courseQuery)

	// If there are any errors
	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecords
		default:
			return nil, err
		}
	}

	// defer closing the course rows
	defer courseRows.Close()

	// loop through course rows
	for courseRows.Next() {

		// temporary Course struct
		var course Course

		// `SELECT courses.course_id, courses.course_code, courses.title, courses.credit, courses.elective
		err := courseRows.Scan(&course.CourseID,
			&course.CourseCode,
			&course.Title,
			&course.Credit,
			&course.Elective)

		// In case of errors
		if err != nil {
			return nil, err
		}

		// Insert into courseRows map
		allCourses[course.CourseID] = course
	}

	// Incase some error occured while scanning the result rows
	if courseRows.Err() != nil {
		return nil, err
	}

	// If all courses count = 0, return immediately from here
	if len(allCourses) == 0 {
		return nil, ErrNoRecords
	}

	// Now that course exists on db, we query for more details

	bookRows, err := m.DB.QueryContext(ctx, bookQuery)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any books listed
		default:
			return extractCourses(allCourses), err

		}
	}

	// Close the bookRows after reading
	defer bookRows.Close()

	// Loop through rows
	for bookRows.Next() {

		// temp book  struct
		var book Book

		var course Course

		// course id to maintain reference
		var courseID int64

		// boolean to check if a book is text book
		var isTextBook bool

		// SELECT courses.course_id,books.book_id, books.title, books.author,
		// books.edition, books.publication, course_books.text_book
		err := bookRows.Scan(&courseID,
			&book.BookID,
			&book.Title,
			&book.Author,
			&book.Edition,
			&book.Publication,
			&isTextBook)

		// If some error occurred
		if err != nil {
			return extractCourses(allCourses), err
		}

		// store the course
		course = allCourses[courseID]

		if isTextBook {

			course.TextBooks = append(course.TextBooks, book)

		} else {
			course.RefBooks = append(course.RefBooks, book)
		}

		// Update that course
		allCourses[courseID] = course

	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = bookRows.Err(); err != nil {
		return extractCourses(allCourses), err
	}

	// Query for teachers
	teacherRows, err := m.DB.QueryContext(ctx, teacherQuery)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any books listed
		default:
			return extractCourses(allCourses), err

		}
	}

	// Defer closing the teacherRows
	defer teacherRows.Close()

	// Loop through teacher rows
	for teacherRows.Next() {

		// temp TaughtBy  struct
		var teacher TaughtBy

		// course id to maintain reference
		var courseID int64

		var course Course

		// SELECT courses.course_id,teachers.teacher_id, teachers.name as teacher
		err := teacherRows.Scan(
			&courseID,
			&teacher.TeacherID,
			&teacher.Name)

		// Incase any err occurred
		if err != nil {
			return extractCourses(allCourses), err
		}

		course = allCourses[courseID]

		// Add the teacher to the course object
		course.TaughtBy = append(course.TaughtBy, teacher)

		// put back the updated course
		allCourses[courseID] = course
	}

	// Incase any error occurred while going through the loop
	if err = teacherRows.Err(); err != nil {
		return extractCourses(allCourses), err
	}

	// Query for offered by
	offeredByRows, err := m.DB.QueryContext(ctx, offeredByQuery)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing and move on
			// since some course may not have any offered by listed
		default:
			return extractCourses(allCourses), err

		}
	}

	// Defer closing the teacherRows
	defer offeredByRows.Close()

	// Loop through rows
	for offeredByRows.Next() {

		// temp offeredBy  struct
		var offeredBy OfferedBy

		// course id to maintain reference
		var courseID int64

		var course Course

		// SELECT courses.course_id, faculties.name as faculty,
		// departments.name as dept, programs.name as program, levels.name as level, program_courses.semester_id as semester
		err := offeredByRows.Scan(
			&courseID,
			&offeredBy.Faculty,
			&offeredBy.Department,
			&offeredBy.Program,
			&offeredBy.Level,
			&offeredBy.Semester)

		// Incase any err occurred
		if err != nil {
			return nil, err
		}

		course = allCourses[courseID]

		// Add the offered by to the course object
		course.OfferedBy = append(course.OfferedBy, offeredBy)

		// Put back the update course
		allCourses[courseID] = course

	}

	// Incase any error occurred while going through the loop
	if err = offeredByRows.Err(); err != nil {
		return extractCourses(allCourses), err
	}

	// Return the courses

	a := extractCourses(allCourses)

	return filterCourses(faculty, department, program, level, semester, &a), nil
}

// Return a slice containing course from the map
func extractCourses(allCourses map[int64]Course) []Course {
	var courses []Course

	// if length is zero, i.e the map is empty
	if len(allCourses) == 0 {
		return nil
	}

	// loop through the courses
	for _, val := range allCourses {
		courses = append(courses, val)
	}

	return courses
}

// filterCourses filters the list of all courses with provided filter values

func filterCourses(faculty, department, program, level string, semester int, allCourses *[]Course) []Course {

	var list []Course

	// Go through course
	for _, val := range *allCourses {

		//  A course may be offered by multiple programs
		temp := val

		// Set the offered by to nil
		temp.OfferedBy = nil
		for _, offer := range val.OfferedBy {

			var add bool = false
			switch {
			case strings.EqualFold(offer.Faculty, faculty):
				add = true
			case strings.EqualFold(offer.Department, department):
				add = true

			case strings.EqualFold(offer.Program, program):
				add = true

			case strings.EqualFold(offer.Level, level):
				add = true

			case offer.Semester == semester:
				add = true

			}

			if add {
				temp.OfferedBy = append(temp.OfferedBy, offer)
			}
		}

		if len(temp.OfferedBy) > 0 {
			list = append(list, temp)
		}

	}

	return list
}
