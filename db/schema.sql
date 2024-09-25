CREATE TABLE `resources` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `uuid` VARCHAR(100) NOT NULL UNIQUE,
    `name` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
) ENGINE=InnoDB;


CREATE TABLE `animal_rankings` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `rank`  INT NOT NULL unique,
    `name` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
) ENGINE=InnoDB;



-- Insert sample data into `resources` table
INSERT INTO `resources` (`uuid`, `name`, `created_at`, `updated_at`) VALUES
('uuid-1', 'Resource 1', '2024-09-25 10:00:00', '2024-09-25 10:00:00'),
('uuid-2', 'Resource 2', '2024-09-25 10:05:00', '2024-09-25 10:05:00'),
('uuid-3', 'Resource 3', '2024-09-25 10:10:00', '2024-09-25 10:10:00'),
('uuid-4', 'Resource 4', '2024-09-25 10:15:00', '2024-09-25 10:15:00'),
('uuid-5', 'Resource 5', '2024-09-25 10:20:00', '2024-09-25 10:20:00'),
('uuid-6', 'Resource 6', '2024-09-25 10:25:00', '2024-09-25 10:25:00'),
('uuid-7', 'Resource 7', '2024-09-25 10:30:00', '2024-09-25 10:30:00'),
('uuid-8', 'Resource 8', '2024-09-25 10:35:00', '2024-09-25 10:35:00'),
('uuid-9', 'Resource 9', '2024-09-25 10:40:00', '2024-09-25 10:40:00'),
('uuid-10', 'Resource 10', '2024-09-25 10:45:00', '2024-09-25 10:45:00'),
('uuid-11', 'Resource 11', '2024-09-25 10:50:00', '2024-09-25 10:50:00'),
('uuid-12', 'Resource 12', '2024-09-25 10:55:00', '2024-09-25 10:55:00'),
('uuid-13', 'Resource 13', '2024-09-25 11:00:00', '2024-09-25 11:00:00'),
('uuid-14', 'Resource 14', '2024-09-25 11:05:00', '2024-09-25 11:05:00'),
('uuid-15', 'Resource 15', '2024-09-25 11:10:00', '2024-09-25 11:10:00'),
('uuid-16', 'Resource 16', '2024-09-25 11:15:00', '2024-09-25 11:15:00'),
('uuid-17', 'Resource 17', '2024-09-25 11:20:00', '2024-09-25 11:20:00'),
('uuid-18', 'Resource 18', '2024-09-25 11:25:00', '2024-09-25 11:25:00'),
('uuid-19', 'Resource 19', '2024-09-25 11:30:00', '2024-09-25 11:30:00'),
('uuid-20', 'Resource 20', '2024-09-25 11:35:00', '2024-09-25 11:35:00');

-- Insert sample data into `animal_rankings` table
INSERT INTO `animal_rankings` (`rank`, `name`, `created_at`, `updated_at`) VALUES
(1, 'Lion', '2024-09-25 10:00:00', '2024-09-25 10:00:00'),
(2, 'Tiger', '2024-09-25 10:05:00', '2024-09-25 10:05:00'),
(3, 'Elephant', '2024-09-25 10:10:00', '2024-09-25 10:10:00'),
(4, 'Leopard', '2024-09-25 10:15:00', '2024-09-25 10:15:00'),
(5, 'Wolf', '2024-09-25 10:20:00', '2024-09-25 10:20:00'),
(6, 'Fox', '2024-09-25 10:25:00', '2024-09-25 10:25:00'),
(7, 'Bear', '2024-09-25 10:30:00', '2024-09-25 10:30:00'),
(8, 'Giraffe', '2024-09-25 10:35:00', '2024-09-25 10:35:00'),
(9, 'Zebra', '2024-09-25 10:40:00', '2024-09-25 10:40:00'),
(10, 'Rhinoceros', '2024-09-25 10:45:00', '2024-09-25 10:45:00'),
(11, 'Hippopotamus', '2024-09-25 10:50:00', '2024-09-25 10:50:00'),
(12, 'Cheetah', '2024-09-25 10:55:00', '2024-09-25 10:55:00'),
(13, 'Jaguar', '2024-09-25 11:00:00', '2024-09-25 11:00:00'),
(14, 'Panda', '2024-09-25 11:05:00', '2024-09-25 11:05:00'),
(15, 'Kangaroo', '2024-09-25 11:10:00', '2024-09-25 11:10:00'),
(16, 'Koala', '2024-09-25 11:15:00', '2024-09-25 11:15:00'),
(17, 'Penguin', '2024-09-25 11:20:00', '2024-09-25 11:20:00'),
(18, 'Ostrich', '2024-09-25 11:25:00', '2024-09-25 11:25:00'),
(19, 'Eagle', '2024-09-25 11:30:00', '2024-09-25 11:30:00'),
(20, 'Falcon', '2024-09-25 11:35:00', '2024-09-25 11:35:00');