# SOE-Backend

This is the backend for the Online Student Portal System developed as a minor Project.

### <i><u> Built With </u></i>

- Go (Gin) for maximum efficiency and lower memory footprint
- PostgreSQL Database

### <i><u> Features </u></i>

- Supports JSON REST API
- Login & Registration (Students, Teachers, Admin)
- Notices (Admin can publish and delete notices)
- View Faculty, Department, Program and other details easily
- Daily Schedule for students, teachers (Admin can publish and delete schedules)
- Lodge Issues (For Students, Teachers)
- Teachers' accounts can viewed as public profiles

### <u><i>Run</i></u>

Please check config_example.toml for configuration details.

```shell

git clone https://github.com/roshanlc/soe-backend

cd soe-backend

cp config_example.toml config.toml

vim config.toml # make neccessary changes

go run cmd/api/*

```

### <u><i>To-do</i></u>

- Test cases
- Cron job to delete expired tokens

### <u><i>Ideas for continuity</i></u>

The plan is to develop a DIY CMS to continue this project as the final project.

1. Id-card generator
2. Examination Form Registration and Verification
3. Marks viewing / uploading
4. Lesson Plans
5. Grade Calculator
6. Certificates
7. Attendance
8. Communication management (between student/teacher, teacher/superuser, student/superuser..)
9. Admission and Applicant Enquiry management
10. Library
