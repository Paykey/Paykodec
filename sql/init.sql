-- (1) 사용자 정보 테이블
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,  -- SERIAL: 자동으로 1씩 증가
    username VARCHAR(50) NOT NULL UNIQUE,   
    password_hash VARCHAR(255) NOT NULL,    -- 비밀번호는 해시 형태로 저장할 예정
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- (2) 라이브러리(폴더) 테이블
CREATE TABLE IF NOT EXISTS libraries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    folder_path TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- (3) 미디어 파일 원본 정보 테이블
CREATE TABLE IF NOT EXISTS media_items (
    id SERIAL PRIMARY KEY,
    library_id INT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    container VARCHAR(20),
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- (4) 사용자별 시청 기록 및 상태 테이블
CREATE TABLE IF NOT EXISTS user_media_data (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_item_id INT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    playback_position_ticks INT DEFAULT 0,
    is_played BOOLEAN DEFAULT FALSE,
    last_played_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, media_item_id)
);