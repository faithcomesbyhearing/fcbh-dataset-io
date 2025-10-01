-- ENGNIVN1DA fileset data for init.db
-- Generated from MySQL database on Wed Oct  1 11:48:00 MDT 2025

-- Fileset data
INSERT INTO bible_filesets (id, hash_id, asset_id, set_type_code, set_size_code, hidden, content_loaded, archived) VALUES ('ENGNIVN1DA', 'bfd3bf1c5beb', 'dbp-prod', 'audio', 'NT', 0, 1, 0);

-- File data (first 6 files for testing)
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589939, 'bfd3bf1c5beb', 'JHN', 1, 'B04___01_John________ENGNIVN1DA.mp3', 3299222, 400);
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589940, 'bfd3bf1c5beb', 'JHN', 2, 'B04___02_John________ENGNIVN1DA.mp3', 1655806, 195);
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589941, 'bfd3bf1c5beb', 'JHN', 3, 'B04___03_John________ENGNIVN1DA.mp3', 2619620, 315);
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589942, 'bfd3bf1c5beb', 'JHN', 4, 'B04___04_John________ENGNIVN1DA.mp3', 3743930, 456);
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589943, 'bfd3bf1c5beb', 'JHN', 5, 'B04___05_John________ENGNIVN1DA.mp3', 3475600, 422);
INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (589944, 'bfd3bf1c5beb', 'JHN', 6, 'B04___06_John________ENGNIVN1DA.mp3', 4452789, 545);

-- Timestamp data (for first file only)
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638008, 589939, '0', NULL, 0, NULL, 0);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638009, 589939, '1', NULL, 10.44, NULL, 1);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638010, 589939, '2', NULL, 13.96, NULL, 2);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638011, 589939, '3', NULL, 15.56, NULL, 3);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638012, 589939, '4', NULL, 22.12, NULL, 4);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638013, 589939, '5', NULL, 27.78, NULL, 5);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638014, 589939, '6', NULL, 33.04, NULL, 6);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638015, 589939, '7', NULL, 38.38, NULL, 7);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638016, 589939, '8', NULL, 45.68, NULL, 8);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638017, 589939, '9', NULL, 50.88, NULL, 9);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638018, 589939, '10', NULL, 56, NULL, 10);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638019, 589939, '11', NULL, 62.58, NULL, 11);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638020, 589939, '12', NULL, 68.16, NULL, 12);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638021, 589939, '13', NULL, 76.18, NULL, 13);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638022, 589939, '14', NULL, 84.88, NULL, 14);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638023, 589939, '15', NULL, 97.36, NULL, 15);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638024, 589939, '16', NULL, 108.16, NULL, 16);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638025, 589939, '17', NULL, 115.58, NULL, 17);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638026, 589939, '18', NULL, 123.44, NULL, 18);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638027, 589939, '19', NULL, 131.04, NULL, 19);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638028, 589939, '20', NULL, 138.22, NULL, 20);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638029, 589939, '21', NULL, 143.64, NULL, 21);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638030, 589939, '22', NULL, 154.28, NULL, 22);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638031, 589939, '23', NULL, 163.32, NULL, 23);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638032, 589939, '24', NULL, 174.4, NULL, 24);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638033, 589939, '25', NULL, 177.2, NULL, 25);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638034, 589939, '26', NULL, 184.38, NULL, 26);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638035, 589939, '27', NULL, 192.12, NULL, 27);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638036, 589939, '28', NULL, 198.56, NULL, 28);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638037, 589939, '29', NULL, 204.92, NULL, 29);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638038, 589939, '30', NULL, 213.32, NULL, 30);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638039, 589939, '31', NULL, 220, NULL, 31);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638040, 589939, '32', NULL, 228.72, NULL, 32);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638041, 589939, '33', NULL, 237.92, NULL, 33);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638042, 589939, '34', NULL, 251.86, NULL, 34);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638043, 589939, '35', NULL, 258.34, NULL, 35);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638044, 589939, '36', NULL, 262.86, NULL, 36);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638045, 589939, '37', NULL, 269, NULL, 37);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638046, 589939, '38', NULL, 272.86, NULL, 38);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638047, 589939, '39', NULL, 283.4, NULL, 39);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638048, 589939, '40', NULL, 294.36, NULL, 40);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638049, 589939, '41', NULL, 301.06, NULL, 41);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638050, 589939, '42', NULL, 308.32, NULL, 42);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638051, 589939, '43', NULL, 321.9, NULL, 43);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638052, 589939, '44', NULL, 329.46, NULL, 44);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638053, 589939, '45', NULL, 333.62, NULL, 45);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638054, 589939, '46', NULL, 345.36, NULL, 46);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638055, 589939, '47', NULL, 351.72, NULL, 47);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638056, 589939, '48', NULL, 360.06, NULL, 48);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638057, 589939, '49', NULL, 369.22, NULL, 49);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638058, 589939, '50', NULL, 378.22, NULL, 50);
INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (638059, 589939, '51', NULL, 387.06, NULL, 51);
