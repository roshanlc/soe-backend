-- contains faculty details
CREATE TABLE IF NOT EXISTS faculties (
	faculty_id serial NOT NULL primary key,
	name text NOT NULL UNIQUE,
	faculty_head text 
);

-- All the departments within a faculty
CREATE TABLE IF NOT EXISTS departments (
	department_id serial NOT NULL primary key,
	name text NOT NULL UNIQUE,
	department_head text, 
	-- heads for any table for now can be null values since theycan be added later.
	faculty_id integer NOT NULL REFERENCES faculties(faculty_id)
);

-- contains levels details (bachelor, master, doctoral ...)
CREATE TABLE IF NOT EXISTS levels (
	level_id serial NOT NULL primary  key,
	name text NOT NULL UNIQUE
);



-- semesters hold all details of semester number
CREATE TABLE IF NOT EXISTS semesters (
	semester_id serial NOT NULL primary key CHECK(semester_id BETWEEN 1 AND 8)
);
-- programs such as comp engg, pharmacy study and also link to the corresponding department
CREATE TABLE IF NOT EXISTS programs (
	program_id serial NOT NULL primary key,
	name text NOT NULL UNIQUE,
	program_head text,
	department_id integer NOT NULL REFERENCES departments(department_id),
	level_id integer NOT NULL REFERENCES levels(level_id)
);


 -- course detail
CREATE TABLE IF NOT EXISTS courses (
	course_id bigserial NOT NULL UNIQUE primary key,
	course_code text NOT NULL UNIQUE,
	title text NOT NULL UNIQUE,
	credit integer NOT NULL DEFAULT 1,
	
	-- is this available only on elective subject
	elective boolean default false
	

	-- the books courses table holds book reference for a course

);



-- this table holds the courses supported by a program
CREATE TABLE IF NOT EXISTS program_courses(
	program_id integer NOT NULL REFERENCES programs(program_id),
	course_id bigint NOT NULL UNIQUE REFERENCES courses(course_id),
	-- semester id should be null for all elective subjects
	semester_id bigint REFERENCES semesters(semester_id),
	
	-- primary key = both the columns
	CONSTRAINT progam_course_pkey PRIMARY KEY (program_id, course_id,semester_id)

);


-- this table holds info about running semesters of each program, dept
CREATE TABLE IF NOT EXISTS running_semesters(
	
	program_id integer NOT NULL REFERENCES programs(program_id),
	semester_id integer NOT NULL UNIQUE REFERENCES semesters(semester_id),
	
	CONSTRAINT running_semesters_pkey PRIMARY KEY(program_id,semester_id)

);


-- this table stores users details

CREATE TABLE IF NOT EXISTS users (
	user_id bigserial NOT NULL primary key,
	email text NOT NULL UNIQUE,
	password text NOT NULL,
	activated boolean NOT NULL DEFAULT false,
	expired boolean NOT NULL DEFAULT false,
	-- version is used in implementing optimistic concurrency control
	version integer NOT NULL DEFAULT 1
);

-- holds logged in users' tokens
CREATE TABLE IF NOT EXISTS tokens (
	hash text PRIMARY KEY,
	user_id bigint NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	expires_at timestamp(0) with time zone NOT NULL DEFAULT (CURRENT_TIMESTAMP(0) + INTERVAL '1 day'),
	scope text NOT NUll
	-- scope can be authentication or authorization
);

-- roles: students, teacher, superuser
-- user: read marks, check time table
-- teacher: upload marks, check time-table, update their public profile pic
-- superuser: add new su, assign time-table, publish notice, activate users

CREATE TABLE IF NOT EXISTS roles (
	role_id serial NOT NULL PRIMARY KEY,
	name text NOT NULL UNIQUE,
	description text -- role description

);

-- users (user id) and their roles
CREATE TABLE IF NOT EXISTS user_roles (
	user_id bigint NOT NULL UNIQUE REFERENCES users(user_id),
	role_id integer NOT NULL REFERENCES roles(role_id),

	-- two foreign keys as primary key
	CONSTRAINT user_roles_pkey PRIMARY KEY(user_id,role_id)
);



-- holds all superusers
CREATE TABLE IF NOT EXISTS superusers(
	superuser_id bigserial NOT NULL PRIMARY KEY,
	name text NOT NULL,
	-- main superuser cannnot be deleted
	main boolean default false, 
	 -- self-referencing
	added_by bigint REFERENCES superusers(superuser_id),

	-- user-details
	-- for main superuser, it should be added manually
	user_id bigint NOT NULL REFERENCES users(user_id)


);

-- website-wide notifications
CREATE TABLE IF NOT EXISTS notices (
	notice_id bigserial NOT NULL PRIMARY KEY,
	title text NOT NULL,
	content  text NOT NULL,
	created_at timestamp(0) with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
	
	media_links text[],
	-- to count the number of modifications
	version integer NOT NULL DEFAULT 1,
	-- who added the notice, no foreign key constraint
	added_by text NOT NULL

	-- add links for pdfs or images
);

-- this table holds student details
CREATE TABLE IF NOT EXISTS students (
	student_id bigserial NOT NULL primary key,
	name text NOT NULL,
	symbol_no bigint UNIQUE NOT NULL,
	pu_regd_no text UNIQUE NOT NULL,
	enrolled_at date NOT NULL DEFAULT NOW(),
	contact_no text,
	version integer NOT NULL DEFAULT 1,
	
	-- program id -> foreign key
	program_id integer NOT NULL REFERENCES programs(program_id),

	-- semester id as the current semester -> foreign key
	semester_id integer NOT NULL REFERENCES semesters(semester_id),	
	-- user_id -> foreign key
	user_id bigint UNIQUE REFERENCES users(user_id)
);



-- this table holds teacher details
-- todo: work on this i.e add more column related to their experience,
--  publcations, education background and their designation(professor, assitant, xyz)
-- profile pictures, 
-- might need to create separate teacher_profiles table only for their public profile
CREATE TABLE IF NOT EXISTS teachers (
	teacher_id bigserial NOT NULL primary key,
	name text NOT NULL,
	joined_at date NOT NULL DEFAULT NOW(),
	contact_no text,
	academics text[] ,
	-- experiences text[],
	-- research_interests text[],
	-- publications text[],

-- teacher profile auto generated from teacher name
	--profile text NOT NULL UNIQUE,

	version integer NOT NULL DEFAULT 1,

	-- user_id -> foreign key
	user_id bigint UNIQUE REFERENCES users(user_id)
);

-- this table holds teachers and the courses they are currently teaching
CREATE TABLE IF NOT EXISTS teacher_courses (

	teacher_id bigint NOT NULL REFERENCES teachers(teacher_id),
	course_id bigint NOT NULL REFERENCES courses(course_id),
	
	-- the course should auto expire at 6 months after adding for a semester
	expires_at DATE NOT NULL DEFAULT (NOW() + INTERVAL '6 months'),
	
	-- two foreign keys  as primary key
	 CONSTRAINT teacher_courses_pkey PRIMARY KEY(teacher_id,course_id)
);


-- books 
CREATE TABLE IF NOT EXISTS books(
	book_id bigserial NOT NULL PRIMARY KEY,
	title text NOT NULL UNIQUE,
	author text NOT NULL,
	edition integer NOT NULL,
	publication text NOT NULL
	
);


-- courses and books relation

CREATE TABLE IF NOT EXISTS course_books(
	course_id bigint NOT NULL UNIQUE REFERENCES courses(course_id),
	book_id bigint NOT NULL REFERENCES books(book_id),

	 --if text_book = false then it is a reference book
	text_book boolean NOT NULL DEFAULT true,

-- foreign keys combination as a primary key
	CONSTRAINT course_books_pkey PRIMARY KEY(course_id,book_id)

);

-- table to store days ("SUNDAY", "MONDAY",...."SATURDAY")
CREATE TABLE IF NOT EXISTS days(
	day varchar(10) NOT NULL PRIMARY KEY
);

-- table to hold time intervals
CREATE TABLE IF NOT EXISTS intervals(
	interval_id serial NOT NULL PRIMARY KEY,
	interval varchar(11) NOT NULL UNIQUE
);


-- daily schedule ( program_id,semester_id,  day, interval, course_id)
CREATE TABLE IF NOT EXISTS day_schedule(
	program_id integer NOT NULL REFERENCES programs(program_id),
	semester_id integer NOT NULL REFERENCES semesters(semester_id),
	day varchar(10) NOT NULL REFERENCES days(day),
	interval_id integer NOT NULL REFERENCES intervals(interval_id),
	course_id bigint NOT NULL REFERENCES courses(course_id),
	description text,

 -- create a primary key based on these three attributes
	CONSTRAINT day_schedule_pkey PRIMARY KEY(program_id,semester_id,day,interval_id,course_id)
);



-- issues 
-- all the issue or complaints registered by users
CREATE TABLE IF NOT EXISTS issues (
issue_id bigserial NOT NULL PRIMARY KEY,
issue text NOT NULL,
user_id bigint NOT NULL REFERENCES users(user_id),
user_role text NOT NULL REFERENCES roles(name),
read boolean NOT NULL DEFAULT false,
created_at timestamp(0) with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP(0)
);


-- teacher profiles
CREATE TABLE IF NOT EXISTS teacher_profiles (
	profile_id text UNIQUE NOT NULL PRIMARY KEY, 
	experiences text[] ,
	publications text[],
	research_interests text[],
	description text, -- for any custom descripton text
	teacher_id bigint UNIQUE NOT NULL REFERENCES teachers(teacher_id)

);