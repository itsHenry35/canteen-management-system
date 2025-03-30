-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL,
    dingtalk_id TEXT
);

-- 学生表
CREATE TABLE IF NOT EXISTS students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    class TEXT NOT NULL,
    dingtalk_id TEXT,
    last_meal_collection_date TIMESTAMP
);


-- 家长-学生关系表
CREATE TABLE IF NOT EXISTS parent_student_relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    UNIQUE(parent_id, student_id)
);

-- 餐表
CREATE TABLE IF NOT EXISTS meals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    selection_start_time TIMESTAMP NOT NULL,
    selection_end_time TIMESTAMP NOT NULL,
    effective_start_date TIMESTAMP NOT NULL,
    effective_end_date TIMESTAMP NOT NULL,
    image_path TEXT NOT NULL
);

-- 选餐记录表
CREATE TABLE IF NOT EXISTS meal_selections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    meal_id INTEGER NOT NULL,
    meal_type TEXT NOT NULL,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (meal_id) REFERENCES meals(id) ON DELETE CASCADE,
    UNIQUE(student_id, meal_id)
);