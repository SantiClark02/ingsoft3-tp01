CREATE DATABASE IF NOT EXISTS backend
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE backend;

CREATE TABLE IF NOT EXISTS usuarios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    rol ENUM('socio','administrador') NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS actividades (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    titulo VARCHAR(100) NOT NULL,
    descripcion TEXT NOT NULL,
    cupo_max SMALLINT NOT NULL,
    imagen_url VARCHAR(255),
    nombre_categoria VARCHAR(50) NOT NULL,
    nombre_instructor VARCHAR(100) NOT NULL,
    email_instructor VARCHAR(150),
    especialidad_instructor VARCHAR(100),
    activo TINYINT(1) NOT NULL DEFAULT 1,
    INDEX idx_actividades_categoria (nombre_categoria)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS horarios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    actividad_id BIGINT NOT NULL,
    dia_semana ENUM('Lun','Mar','Mie','Jue','Vie','Sab','Dom') NOT NULL,
    hora_inicio TIME NOT NULL,
    duracion_min SMALLINT NOT NULL,
    CONSTRAINT fk_horarios_actividades FOREIGN KEY (actividad_id)
        REFERENCES actividades(id)
        ON DELETE CASCADE,
    INDEX idx_horarios_actividad (actividad_id),
    INDEX idx_horarios_dia_semana (dia_semana)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS inscripciones (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    usuario_id BIGINT NOT NULL,
    horario_id BIGINT NOT NULL,
    fecha_inscripcion DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(usuario_id, horario_id),
    CONSTRAINT fk_inscripciones_usuarios FOREIGN KEY (usuario_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_inscripciones_horarios FOREIGN KEY (horario_id)
        REFERENCES horarios(id)
        ON DELETE CASCADE,
    INDEX idx_inscripciones_usuario (usuario_id),
    INDEX idx_inscripciones_horario (horario_id)
) ENGINE=InnoDB;

INSERT INTO usuarios (nombre, email, password_hash, rol) VALUES
  ('emiliano', 'emiliano@mail.com', '$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'socio'),
  ('maria',    'maria@mail.com',    '$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'socio'),
  ('admin',    'admin@gimnasio.com','$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'administrador');

INSERT INTO actividades 
  (titulo, descripcion, cupo_max, imagen_url, nombre_categoria, nombre_instructor, email_instructor, especialidad_instructor, activo)
VALUES
  ('Zumba Avanzado',
   'Clase intensa de Zumba combinando ritmos latinos y cardio.',
   20,
   'https://landing-pages-smart.s3.amazonaws.com/modalidades-treinos/es/zumba/cover.png',
   'Baile',
   'Florencia Ruiz',
   'florencia@instructores.com',
   'Zumba certificada',
   TRUE),
  ('Funcional Básico',
   'Entrenamiento funcional para todos los niveles, usa peso corporal.',
   15,
   'https://www.fhtrainer.com/wp-content/uploads/2020/11/entrenamiento-de-alta-intensidad-HIIT-Blog.jpg',
   'Fitness',
   'Carlos Pérez',
   'carlos@instructores.com',
   'Entrenador Personal',
   TRUE),
  ('Pilates Suave',
   'Rutina suave de Pilates para mejorar la flexibilidad y el core.',
   10,
   'https://ignisfisioterapiagirona.cat/wp-content/uploads/2020/01/pilates-girona.jpg',
   'Pilates',
   'Ana Gómez',
   'Ana@gmail.com',
   'Entrenador Personal',
   TRUE);

INSERT INTO horarios (actividad_id, dia_semana, hora_inicio, duracion_min) VALUES
  (1, 'Lun', '18:00:00', 60),
  (1, 'Mie', '19:00:00', 60),
  (1, 'Vie', '17:00:00', 60),
  (2, 'Mar', '18:30:00', 45),
  (2, 'Jue', '19:30:00', 45),
  (3, 'Lun', '10:00:00', 50),
  (3, 'Mie', '10:00:00', 50);

INSERT INTO inscripciones (usuario_id, horario_id) VALUES
  (1, 1),
  (1, 4),
  (2, 6);

INSERT INTO usuarios (nombre, email, password_hash, rol) VALUES
  ('juanperez',  'juan@mail.com',  '$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'socio'),
  ('luisa',      'luisa@mail.com', '$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'socio'),
  ('pabloh',     'pablo@mail.com', '$2b$12$Kzai13XwycEQFCgCySDAjeaPtRTmHpFaCJzJGISytzlskjFkbzhvO', 'socio');

INSERT INTO actividades 
  (titulo, descripcion, cupo_max, imagen_url, nombre_categoria, nombre_instructor, email_instructor, especialidad_instructor, activo)
VALUES
  ('Zumba Intermedio',
   'Clase de Zumba para nivel intermedio, con coreografía desafiante.',
   25,
   'https://statics-cuidateplus.marca.com/cms/images/zumba.jpg',
   'Baile',
   'María López',
   'maria@instructores.com',
   'Ritmos Latinos',
   TRUE),
  ('Funcional Intensivo',
   'Entrenamiento funcional de alta intensidad para quemar calorías.',
   20,
   'https://live.staticflickr.com/3800/14088400880_b9f70a052c_b.jpg',
   'Fitness',
   'Carlos Pérez',
   'carlos@instructores.com',
   'CrossFit Trainer',
   TRUE),
  ('Calistenia Básica',
   'Introducción a la calistenia: uso del peso corporal.',
   15,
   'https://www.calistenia.net/wp-content/uploads/2020/08/entrenamiento.png',
   'Calistenia',
   'Ana Gómez',
   NULL,
   NULL,
   TRUE),
  ('Spinning Avanzado',
   'Clase de ciclismo indoor con intervalos de alta resistencia.',
   30,
   'https://i.blogs.es/1a2e41/istock-1134717663/1366_2000.jpeg',
   'Ciclismo',
   'Roberto Díaz',
   'roberto@instructores.com',
   'Instructor Spinning',
   TRUE);

INSERT INTO horarios (actividad_id, dia_semana, hora_inicio, duracion_min) VALUES
  (4, 'Mar', '18:00:00', 60),
  (4, 'Jue', '19:00:00', 60),
  (5, 'Lun', '20:00:00', 50),
  (5, 'Mie', '20:00:00', 50),
  (5, 'Vie', '20:00:00', 50),
  (6, 'Mar', '10:00:00', 45),
  (6, 'Jue', '10:00:00', 45),
  (7, 'Mie', '07:00:00', 45),
  (7, 'Sab', '09:00:00', 50);
