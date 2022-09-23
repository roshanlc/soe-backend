-- insert into roles
INSERT INTO roles
VALUES (default,'student','Student role'),
(default,'teacher','Teacher role'),
(default,'superuser','Superuser role');

-- Insert into levels
INSERT INTO levels VALUES(default, 'Masters');
INSERT INTO levels VALUES(default, 'Bachelors');

-- Insert into faculties
INSERT INTO faculties VALUES(default, 'Faculty of Science and Technology', '');
 

-- Insert into departments
INSERT INTO departments
VALUES 
( default, 'Department of Computer and Software Engineering','',(SELECT faculty_id FROM faculties WHERE name = 'Faculty of Science and Technology'));



-- insert into semesters
INSERT INTO semesters 
VALUES (1),(2),(3),(4),(5),(6),(7),(8);

-- Insert into programs
INSERT INTO programs 
VALUES 
(default, 'Computer Engineering', '',
(SELECT department_id FROM departments WHERE name = 'Department of Computer and Software Engineering'),
(SELECT level_id from levels WHERE name = 'Bachelors'));



-- insert into running_semesters
INSERT INTO running_semesters
VALUES ((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6);


-- insert into users (password=Thisistheway)
INSERT INTO users 
VALUES 
(default,'student1@pu.edu.np','$2a$12$gJjLl6GS4IuYIEBVPVZhP.tBPQObiuv9ddtdWx/FxeP4JJbcMhH5G','t','f',default),
(default,'student2@pu.edu.np','$2a$12$gJjLl6GS4IuYIEBVPVZhP.tBPQObiuv9ddtdWx/FxeP4JJbcMhH5G','t','f',default),
(default,'teacher1@pu.edu.np','$2a$12$gJjLl6GS4IuYIEBVPVZhP.tBPQObiuv9ddtdWx/FxeP4JJbcMhH5G','t','f',default),
(default,'teacher2@pu.edu.np','$2a$12$gJjLl6GS4IuYIEBVPVZhP.tBPQObiuv9ddtdWx/FxeP4JJbcMhH5G','t','f',default),
(default,'admin1@pu.edu.np','$2a$12$gJjLl6GS4IuYIEBVPVZhP.tBPQObiuv9ddtdWx/FxeP4JJbcMhH5G','t','f',default);

-- insert into students

INSERT INTO students
VALUES 
(default,'John Doe',19078979,'2018-36-56',default,'+977-65465659',default,(SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,
(SELECT user_id FROM users WHERE email = 'student1@pu.edu.np')),
(default,'Ronny Rox',19078969,'2018-69-56',default,'+977-65656',default,(SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,
(SELECT user_id FROM users WHERE email = 'student2@pu.edu.np'));

-- insert into teachers

INSERT INTO teachers
VALUES 
(default,'Balen Shah',default,'+977-165656','{"Structural Engineering","Civil Engineering"}',default,(SELECT user_id FROM users WHERE email = 'teacher1@pu.edu.np')),
(default,'Anand Gandhi',default,'+977-1898956','{"Filmgraphy","Philosohpy", "Artificial Intelligence"}',default,(SELECT user_id FROM users WHERE email = 'teacher2@pu.edu.np'));

INSERT INTO teacher_profiles
VALUES
('balend_shah', '{"Teaching for two years","Engineer for 4 years"}','{}','{"Development of ktm as a mayor"}','I am a rapper.',(SELECT teacher_id FROM teachers WHERE name = 'Balen Shah')),
('anand_gandhi','{"Filmmaker for twenty years","Writer for 20+ years"}','{}','{"Development of ktm as a mayor"}','Film-maker, Philosopher, Artist, Biologist.',(SELECT teacher_id FROM teachers WHERE name = 'Anand Gandhi'));


-- insert into superusers
INSERT INTO superusers
VALUES (1,'Naval','t',1,(SELECT user_id FROM users WHERE email = 'admin1@pu.edu.np'));


-- insert into user_roles
INSERT INTO user_roles
VALUES 
((SELECT user_id FROM users WHERE email = 'student1@pu.edu.np'),(SELECT role_id FROM roles where name ='student')),
((SELECT user_id FROM users WHERE email = 'student2@pu.edu.np'),(SELECT role_id FROM roles where name ='student')),
((SELECT user_id FROM users WHERE email = 'teacher1@pu.edu.np'),(SELECT role_id FROM roles where name ='teacher')),
((SELECT user_id FROM users WHERE email = 'teacher2@pu.edu.np'),(SELECT role_id FROM roles where name ='teacher')),
((SELECT user_id FROM users WHERE email = 'admin1@pu.edu.np'),(SELECT role_id FROM roles where name ='superuser'));

-- insert into courses
INSERT INTO courses
VALUES 
(default,'CMP115','Object Oriented Programming C++',3,'f'),
(default,'PHY111','Physics',4,'f'),
(default,'MEC130','Applied Mechanics',3,'f'),
(default,'CMP350','Simulation and Modeling',3,'f'),
(default,'CMP320','Object Oriented Software Engineering',3,'f'),
(default,'CMP390','Advanced Web Technology',3,'t');

-- insert into teacher_courses
 INSERT INTO teacher_courses 
 VALUES ((SELECT teacher_id FROM teachers WHERE name='Balen Shah'),(SELECT course_id FROM courses WHERE course_code = 'CMP115'),default),
 ((SELECT teacher_id FROM teachers WHERE name='Anand Gandhi'),(SELECT course_id FROM courses WHERE course_code = 'CMP390'),default),
 ((SELECT teacher_id FROM teachers WHERE name='Balen Shah'),(SELECT course_id FROM courses WHERE course_code = 'MEC130'),default),
 ((SELECT teacher_id FROM teachers WHERE name='Anand Gandhi'),(SELECT course_id FROM courses WHERE course_code = 'CMP350'),default),
 ((SELECT teacher_id FROM teachers WHERE name='Balen Shah'),(SELECT course_id FROM courses WHERE course_code = 'CMP320'),default);
 
 
-- insert into books
INSERT INTO books
VALUES (default,'C++ Programming','Bjarne Strastoup',2,'Some XYZ'),
(default,'Quantum Physics', 'Feymann',2,'Some XYZ'),
(default,'Software Engineering Guide','Some author',4,'XYZ Some'),
(default,'Programming The WWW','Robert',8,'Pearson Publications');

-- insert into course_books
INSERT INTO course_books 
VALUES
((SELECT course_id FROM courses WHERE course_code = 'CMP115'),(SELECT book_id FROM books WHERE title= 'C++ Programming'),'t'),
((SELECT course_id FROM courses WHERE course_code = 'PHY111'),(SELECT book_id FROM books WHERE title= 'Quantum Physics'),'t'),
((SELECT course_id FROM courses WHERE course_code = 'CMP320'),(SELECT book_id FROM books WHERE title= 'Software Engineering Guide'),'t'),
((SELECT course_id FROM courses WHERE course_code = 'CMP390'),(SELECT book_id FROM books WHERE title= 'Programming The WWW'),'t');

-- insert into program_courses
INSERT INTO program_courses 
VALUES 
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'CMP115'),2),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'PHY111'),2),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'MEC130'),2),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'CMP350'),6),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'CMP320'),6),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),(SELECT course_id FROM courses WHERE course_code = 'CMP390'),6);

-- insert into notices

-- insert into days
INSERT INTO days VALUES 
('SUNDAY'),
('MONDAY'),
('TUESDAY'),
('WEDNESDAY'),
('THURSDAY'),
('FRIDAY'),
('SATURDAY');


-- insert into intervals
INSERT INTO intervals VALUES 
(default,'10:15-11:05'),
(default,'11:05-11:55'),
(default,'11:55-12:45'),
(default,'12:45-13:35'),
(default,'13:35-14:05'),
(default,'14:05-14:55'),
(default,'14:55-15:45'),
(default,'15:55-16:45');


-- daily schedule ( program_id,semester_id,  day, interval, course_id)
INSERT INTO day_schedule VALUES 

-- for second semester
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'SUNDAY',
(SELECT interval_id FROM intervals WHERE interval = '10:15-11:05'),(SELECT course_id FROM courses WHERE course_code = 'CMP115')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'MONDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'CMP115')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'TUESDAY',
(SELECT interval_id FROM intervals WHERE interval = '12:45-13:35'),(SELECT course_id FROM courses WHERE course_code = 'CMP115')),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'MONDAY',
(SELECT interval_id FROM intervals WHERE interval = '10:15-11:05'),(SELECT course_id FROM courses WHERE course_code = 'MEC130')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'SUNDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'MEC130')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),2,'TUESDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'MEC130')),

-- for sixth semester
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'WEDNESDAY',
(SELECT interval_id FROM intervals WHERE interval = '10:15-11:05'),(SELECT course_id FROM courses WHERE course_code = 'CMP350')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'THURSDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'CMP350')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'FRIDAY',
(SELECT interval_id FROM intervals WHERE interval = '12:45-13:35'),(SELECT course_id FROM courses WHERE course_code = 'CMP350')),
((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'THURSDAY',
(SELECT interval_id FROM intervals WHERE interval = '10:15-11:05'),(SELECT course_id FROM courses WHERE course_code = 'CMP320')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'WEDNESDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'CMP320')),

((SELECT program_id FROM programs WHERE name = 'Computer Engineering'),6,'FRIDAY',
(SELECT interval_id FROM intervals WHERE interval = '11:55-12:45'),(SELECT course_id FROM courses WHERE course_code = 'CMP320'))
;



-- issues
--  issue_id |issue| user_id | user_role | read |created_at
INSERT INTO issues VALUES 
(default, 'This a issue text',(SELECT user_id FROM users WHERE email = 'student1@pu.edu.np'),'student','f',default),
(default, 'This a issue text',(SELECT user_id FROM users WHERE email = 'teacher1@pu.edu.np'),'teacher','f',default);


-- insert a sample notice

INSERT INTO notices VALUES (default,
'This is a demo notice title',
'Nepal adopted the multi-university concept in 1983. The idea of Pokhara University (PU) was conceived in 1986; however, it was established only in 1997 under the Pokhara University Act, 1997. The Incumbent Honorable Prime Minister and the Honorable Minister for Education of the Federal Democratic Republic Nepal are the Chancellor and the Pro-Chancellor, respectively.  The Chancellor appoints the Vice Chancellor, the principal executive officer of the university. The Registrar is designated to assist him/her in financial management and general administration. A non-profit autonomous institution, PU is partly funded by the Government of Nepal and partly by revenues from its students and affiliated colleges.',
default,
'{}',
default,
'naval');