# AudioHLS.py

# given input bibleid[s][,type], create corresponding HLS filesets in DB (no extra files in S3 are needed)
#
## approach: INSERT/SELECT as we go, risking partial loads and paying sequential overhead, but simplifying bookkeeping
##           support one or many input bibleid's (btw could invoke one bibleid per aws lambda to parallelize large sets)
## feedback: one line per bibleid/fileset, with letters indicating book being worked (or . if book not in fileset), newline at end
# foreach bibleid (list of input bibleid[/fileset])
#   foreach type_fileset (audio drama,nondrama) # DISTINCT(fileset_id) FROM bible_filesets WHERE set_type_code IN ('audio_drama','audio') AND bitrate=64k
#       # wrinkle: use only C fileset if N and O also exists (duplicates!?  verify C's work on live.bible.is and new bible.is app)
#     print("bibleid/type_fileset:", end=" ") # so can tell what program is working on
#     INSERT INTO bible_filesets (eg ENGESVN2SA)
#     stream_hashid = SELECT FROM bible_filesets WHERE id = ENGESVN2SA
#     local_mp3s = get list of local mp3's (download from s3 if needed)
#        # enforce cannon book/chapter order in mp3s or sql query, and foreach on that (for progress monitoring)
#     foreach chapter # SELECT file_name FROM bible_files WHERE hash_id=hash_id
#       if new book, print first letter of book (or . for missing books in cannon-order sequence)
#       foreach bitrate_hashid # DISTINCT(description) FROM bible_fileset_tags WHERE hash_id=hashid AND name='bitrate'
#         # do work
#         verify chapter is available in mp3s (eg B01___01_Matthew_____ENGESVN2DA.mp3)
#         use ffprobe to determine bitrate
#         [bytes,offsets,durations] = use ffprobe to determine verse offset, bytes, and duration
#         # save work
#         INSERT INTO bible_files (stream_hashid, eg B01___01_Matthew_____ENGESVN2DA.m3u8)
#         fileid = SELECT FROM bible_files (what we inserted above)
#         INSERT INTO bible_files_stream_bandwidths (fileid, eg B01___01_Matthew_____ENGESVN2DA-bitrate.m3u8, bitrate)
#         bwid = SELECT FROM bible_files_stream_bandwidths (what we inserted above)
#         INSERT INTO bible_files_stream_segments (bwid, timestamps.id, [bytes,offsets,durations])

import sys
import os
import re
import pymysql
import boto3
import subprocess
from subprocess import Popen, PIPE
import hashlib
from Config import *

## insert into bible_fileset_types (set_type_code, name) values ('audio_stream','HLS Audio Stream');
## insert into bible_fileset_types (set_type_code, name) values ('audio_drama_stream', 'HLS Audio Stream Drama');

## Lambda entry point
def handler(event, context):
	print("This method will be used to start program from an AWS Lambda event")
    #for record in event['Records']:
    #    bucket = record['s3']['bucket']['name']
    #    key = unquote_plus(record['s3']['object']['key'])
    #    download_path = '/tmp/{}{}'.format(uuid.uuid4(), key)
    #    s3_client.download_file(bucket, key, download_path)
    #    # pass to validateArguments
    #    # call processFilesetId and processBibleId as needed

class AudioHLS:


	def __init__(self, config):
		self.adapter = AudioHLSAdapter(config) # This class is found later in this file
		session = boto3.Session(profile_name=config.s3_aws_profile)
		self.s3Client = session.client("s3")
		self.bitrateRegex = re.compile(r"bit_rate=([0-9]+)")
		self.timesRegex = re.compile(r"best_effort_timestamp_time=([0-9.]+)\|pkt_pos=([0-9]+)")
		self.directory_audio_hls = config.directory_audio_hls
		self.audio_hls_duration_limit = config.audio_hls_duration_limit


	def close(self):
		self.adapter.close()


	## Command line entry point
	def processCommandLine(self):
		if len(sys.argv) < 3:
			print("ERROR: Enter config_profile and any number of bibleids or bibleid/filesetids to process on command line.")
			sys.exit()
		arguments = sys.argv[2:]
		self.validateArguments(arguments)
		for arg in arguments:
			parts = arg.split("/")
			if len(parts) > 1:
				self.processFilesetId(parts[0], parts[1])
			else:
				self.processBibleId(arg)
		self.close()


	## This is used by both processCommandLine and lambda handler
	def validateArguments(self, arguments):
		errorCount = 0
		for arg in arguments:
			parts = arg.split("/")
			if len(parts) == 1:
				if self.adapter.checkBibleId(arg) < 1:
					errorCount += 1
					print("ERROR: bibleid: %s does not exist." % (arg))
				elif self.adapter.checkBibleTimestamps(arg) < 1:
					errorCount += 1
					print("ERROR: bibleid: %s has no timestamp data." % (arg))
			elif len(parts) == 2:
				if len(parts[1]) != 10:
					errorCount += 1
					print("ERROR: filesetid must be 10 characters.")
				if parts[1][8:10] != "DA":
					errorCount += 1
					print("ERROR: filesetid: %s must have DA in 8,9 position.")
				elif self.adapter.checkFilesetId(parts[0], parts[1]) < 1:
					errorCount += 1
					print("ERROR: bibleid/filesetid pair: %s does not exist." % (arg))
				elif self.adapter.checkFilesetTimestamps(parts[1]) < 1:
					errorCount += 1
					print("ERROR: filesetid: %s has no timestamp data." % (arg))
			else:
				errorCount += 1
				print("ERROR: argument: %s is invalid and was not processed." % (arg))
		if errorCount > 0:
			sys.exit()


	def processBibleId(self, bibleId):
		filesetList = self.adapter.selectFilesetIds(bibleId)
		for filesetId in filesetList:
			self.processFilesetId(bibleId, filesetId)


	def processFilesetId(self, bibleId, origFilesetId):
		self.adapter.createAudioHLSWork(origFilesetId)
		bitrateMap = self.adapter.getBitrateMap() ## bitrate: (filesetId, hashId)

		## check for the presence of timestamp for each fileset
		countError = False
		for bitrate, (bitrateFilesetId, bitrateHashId) in bitrateMap.items():
			count = self.adapter.checkFilesetTimestamps(bitrateFilesetId)
			if count < 10:
				countError = True
				print("ERROR: filesetId %s has %d rows of timestamp data." % (bitrateFilesetId, count))
		if countError:
			sys.exit()

		if bitrateMap.get("64") == None:
			self.adapter.dropAudioHLSWork()
			return
		## Check for dropped files
		(filesetId, origHashId) = bitrateMap["64"]
		sql = ("SELECT distinct file_name FROM bible_files WHERE hash_id = %s"
			" AND id NOT IN (SELECT file_id FROM audio_hls_work)")
		missingList = self.adapter.selectList(sql, (origHashId,))
		for missing in missingList:
			print("WARN: %s was dropped, because it was not present in all bitrates." % (missing))

		## check for duration mismatch
		sql = ("SELECT file_name, MIN(duration), MAX(duration), count(*) FROM audio_hls_work"
			" WHERE hash_id IN (%s, %s) GROUP BY file_name")
		toDelete = []
		for bitrate in bitrateMap.keys():
			if bitrate != "64":
				(bitrateFilesetId, bitrateHashId) = bitrateMap[bitrate]
				resultSet = self.adapter.select(sql, (bitrateHashId, origHashId))
				for row in resultSet:
					minDuration = row[1]
					maxDuration = row[2]
					if (maxDuration - minDuration) > self.audio_hls_duration_limit:
						fileName = row[0]
						toDelete.append((bitrateFilesetId, fileName))
						print("WARN: %s/%s/%s was dropped, because of duration difference %d to %d" % (bibleId, bitrateFilesetId, fileName, minDuration, maxDuration))

		## Remove those that fail duration match
		if len(toDelete) > 0:
			for aDelete in toDelete:
				self.adapter.execute("DELETE FROM audio_hls_work WHERE fileset_id = %s AND file_name = %s", aDelete)
		#self.adapter.printAudioHLSWork()

		print("%s/%s: " % (bibleId, origFilesetId), end="", flush=True)
		fileset = self.adapter.selectFileset(origHashId)
		filesetId = origFilesetId[0:8] + "SA"
		assetId = fileset[0]
		if fileset[1] == "audio":
			setTypeCode = "audio_stream"
		elif fileset[1] == "audio_drama":
			setTypeCode = "audio_drama_stream"
		else:
			raise("Invalid set_type_code %s" % (fileset[1]))
		setSizeCode = fileset[2]
		hashId = self.hashId(assetId, filesetId, setTypeCode)

		## Select all needed data before transaction starts
		files = self.adapter.selectBibleFiles(origHashId) # list of files
		allFilesMap = self.adapter.selectAllBitrateBibleFiles() # filesetId:book_id:chapter:verse : file_name
		timestampMap = self.adapter.selectTimestamps() # filesetId:book:chapter : [(timestamp_id, timestamp)]

		origFilename = None
		try:
			## Insert SA Filesets and Tags
			self.adapter.beginFilesetInsertTran()
			self.adapter.deleteSAFilesetFiles(hashId)
			self.adapter.insertFileset((hashId, filesetId, assetId, setTypeCode, setSizeCode))
			self.adapter.copyFilesetTables(hashId, origHashId)

			currBook = None
			for file in files:
				#(753506, 'B01___01_Matthew_____ENGESVN2DA.mp3', 'MAT', 1, None, 1, None, 1724323, 210)
				origFileId = file[0]
				if currBook != file[2]:
					currBook = file[2]
					print(currBook[0], end="", flush=True)
				origFilename = file[1]
				filename = origFilename.split(".")[0][:-2] + "SA.m3u8"
				values = (filename,) + file[2:]
				chapterStart = file[3]
				verseStart = file[5]
				self.adapter.insertFiles(values)
				self.adapter.copyFileTags(origFileId)

				for bitrate in bitrateMap.keys():
					(bitrateFilesetId, bitrateHashId) = bitrateMap[bitrate]
					key = "%s:%s:%d:%d" % (bitrateFilesetId, currBook, chapterStart, verseStart)
					bitrateFilename = allFilesMap.get(key)
					if bitrateFilename != None:
						mp3FilePath = self.getMP3File(assetId, bibleId, bitrateFilesetId, bitrateFilename)
						realBitrate = self.getBitrate(mp3FilePath)
						outputFilename = filename.split(".")[0] + "-" + bitrate + "kbs.m3u8"
						self.adapter.insertBandwidth((outputFilename, realBitrate))

						key = "%s:%s:%s" % (bitrateFilesetId, file[2], file[3]) # filesetId:book:chapter
						timestamps = timestampMap[key]
						for segment in self.getBoundaries(mp3FilePath, timestamps):
							self.adapter.addSegment(segment)
						self.adapter.insertSegments()
			self.adapter.copyTimestamps(hashId, origHashId)
			self.adapter.commitFilesetInsertTran()
			self.adapter.dropAudioHLSWork()
			print("", end="\n", flush=True)
		except Exception as error:
			print(("\nERROR: at %s %s" % (origFilename, error)), flush=True)
			self.adapter.rollbackFilesetInsertTran()
			raise error ## Comment out for production


	def hashId(self, bucket, filesetId, typeCode):
		md5 = hashlib.md5()
		md5.update(filesetId.encode("latin1"))
		md5.update(bucket.encode("latin1"))
		md5.update(typeCode.encode("latin1"))
		hash_id = md5.hexdigest()
		return hash_id[:12]


	def getMP3File(self, s3Bucket, bibleId, filesetId, filename):
		s3Key = "audio/%s/%s/%s" % (bibleId, filesetId, filename)
		filepath = self.directory_audio_hls + os.sep + s3Key
		if not os.access(filepath, os.R_OK):
			#print("Must download %s" % (s3Key))
			directory = filepath[:filepath.rfind("/")]
			if not os.access(directory, os.R_OK):
				os.makedirs(directory)
			self.s3Client.download_file(s3Bucket, s3Key, filepath)
		return filepath


	def getBitrate(self, file):
		cmd = 'ffprobe -select_streams a -v error -show_format ' + file + ' | grep bit_rate'
		response = subprocess.run(cmd, shell=True, capture_output=True)
		result = self.bitrateRegex.search(str(response))
		if result != None:
			return result.group(1)
		else:
			raise Exception("ffprobe for bitrate failed for %s" % (file))


	def getBoundaries(self, file, times):
		cmd = 'ffprobe -show_frames -select_streams a -of compact -show_entries frame=best_effort_timestamp_time,pkt_pos ' + file
		pipe = Popen(cmd, shell=True, stdout=PIPE, stderr=PIPE)
		i = prevtime = prevpos = time = pos = 0
		(timestamp_id, bound) = times[i]
		hasResults = False
		for line in pipe.stdout:
			hasResults = True
			tm = self.timesRegex.search(str(line))
			time = float(tm.group(1))
			pos  = int(tm.group(2))
			if (time >= bound):
				duration = time - prevtime
				nbytes = pos - prevpos
				yield (timestamp_id, round(duration, 4), prevpos, nbytes)
				prevtime, prevpos = time, pos
				if (i+1 != len(times)):
					i += 1
					(timestamp_id, bound) = times[i]
				else:
					bound = 99999999 # search to end of pipe
		if not hasResults:
			raise Exception("ffprobe failed to return position data.")
		#duration = time - prevtime
		#nbytes = pos - prevpos
		#yield (timestamp_id, round(duration, 4), prevpos, nbytes)



class AudioHLSAdapter:

	def __init__(self, config):
		self.db = pymysql.connect(host = config.database_host,
                             		user = config.database_user,
                             		password = config.database_passwd,
                             		db = config.database_db_name,
                             		port = config.database_port,
                             		charset = 'utf8mb4',
                             		cursorclass = pymysql.cursors.Cursor)
		self.tranCursor = None
		self.currHashId = None
		self.currFileId = None
		self.currBandwidthId = None
		self.segments = []
		self.sqlLog = open("AudioHLSSQL.log", mode="wt")


	def close(self):
		if self.db != None:
			self.db.close()
			self.db = None
		if self.sqlLog != None:
			self.sqlLog.close()
			self.sqlLog = None


	## Query finds audio filesets without HLS where timestamp data is available
	## Limitation: It cannot distinquish between a fileset that has a complete HLS and one with a little
	## This is not being used
	def findFilesetIdNeedHLS(self):
		sql = ("SELECT DISTINCT bs.id"
			" FROM bible_file_timestamps bft, bible_files bf, bible_filesets bs"
			" WHERE bft.bible_file_id = bf.id AND bf.hash_id = bs.hash_id"
			" AND bs.id NOT IN (SELECT CONCAT(LEFT(id,8),'DA',MID(id,11,2))"
			" FROM bible_filesets WHERE set_type_code IN ('audio_stream', 'audio_drama_stream')")
		return self.selectList(sql, None)


	def checkBibleId(self, bibleId):
		sql = "SELECT count(*) FROM bible_fileset_connections WHERE bible_id=%s"
		return self.selectScalar(sql, (bibleId,))


	def checkFilesetId(self, bibleId, filesetId):
		sql = ("SELECT count(*) FROM bible_fileset_connections bfc, bible_filesets bf"
			" WHERE bfc.bible_id=%s AND bf.id=%s")
		return self.selectScalar(sql, (bibleId, filesetId))


	def checkBibleTimestamps(self, bibleId):
		sql = ("SELECT count(*) FROM bible_file_timestamps ft, bible_files bf, bible_filesets bfs,"
			" bible_fileset_connections bfc"
			" WHERE bf.id=ft.bible_file_id AND bfs.hash_id=bf.hash_id AND bfc.hash_id=bfs.hash_id"
			" AND bfc.bible_id=%s")
		return self.selectScalar(sql, (bibleId,))


	## Returns the number of rows of timestamp data available for a filesetId
	def checkFilesetTimestamps(self, filesetId):
		sql = ("SELECT count(*) FROM bible_file_timestamps ft, bible_files bf, bible_filesets bfs"
			" WHERE bf.id=ft.bible_file_id AND bfs.hash_id=bf.hash_id AND bfs.id=%s")
		return self.selectScalar(sql, (filesetId,))


	## Returns the audio fileset_ids for a bible_id
	def selectFilesetIds(self, bibleId):
		sql = ("SELECT bf.id FROM bible_filesets bf, bible_fileset_connections bfc"
			" WHERE bf.hash_id=bfc.hash_id AND bfc.bible_id=%s"
			" AND bf.set_type_code IN ('audio', 'audio_drama')")
		return self.selectList(sql, (bibleId,))

	## Creates a temporary table that contains the matching DA files that will be used to generate the SA fileset
	def createAudioHLSWork(self, filesetId):
		sql2 = ("CREATE TEMPORARY TABLE audio_hls_work(fileset_id VARCHAR(16) NOT NULL, hash_id VARCHAR(16) NOT NULL,"
			" file_id int NOT NULL, file_name VARCHAR(128) NOT NULL, duration int NULL)"
			" SELECT bs2.id AS fileset_id, bs2.hash_id, bf2.id AS file_id, bf2.file_name, bf2.duration"
			" FROM bible_filesets bs1, bible_filesets bs2, bible_files bf1, bible_files bf2"
			" WHERE bs1.id = LEFT(bs2.id, 10)"
			" AND bs1.set_type_code = bs2.set_type_code"
			" AND bs1.id = %s"
			" AND bs1.hash_id = bf1.hash_id"
			" AND bs2.hash_id = bf2.hash_id"
			" AND bf1.book_id = bf2.book_id"
			" AND bf1.chapter_start = bf2.chapter_start"
			" AND bf1.verse_start = bf2.verse_start")
		self.execute(sql2, (filesetId,))


	def dropAudioHLSWork(self):
		self.execute("DROP TEMPORARY TABLE IF EXISTS audio_hls_work", ())


	## Debug method to display contents of audio_hls_work
	def printAudioHLSWork(self):
		resultSet = self.select("SELECT * FROM audio_hls_work ORDER BY file_name", ())
		for row in resultSet:
			print(row)

	def getBitrateMap(self):
		resultSet = self.select("SELECT distinct fileset_id, hash_id FROM audio_hls_work", ())
		result = {}
		for row in resultSet:
			filesetId = row[0]
			hashId = row[1]
			bitrate = filesetId[10:]
			if bitrate == "":
				bitrate = "64"
			result[bitrate] = (filesetId, hashId)
		return result


	def selectFileset(self, hashId):
		sql = "SELECT asset_id, set_type_code, set_size_code, hash_id FROM bible_filesets WHERE hash_id=%s"
		return self.selectRow(sql, (hashId,))


    ## Returns the common DA files associated with a fileset
	def selectBibleFiles(self, hashId):
		sql = ("SELECT id, file_name, book_id, chapter_start, chapter_end, verse_start,"
				" verse_end, file_size, duration"
				" FROM bible_files WHERE id IN"
				" (SELECT file_id FROM audio_hls_work WHERE hash_id = %s) ORDER BY file_name")
		return self.select(sql, (hashId,))


	## Returns a Map of all bible files files being processed: fileset_id/book_id/chapter/verse: filename
	def selectAllBitrateBibleFiles(self):
		sql = ("SELECT w.fileset_id, bf.book_id, bf.chapter_start, bf.verse_start, w.file_name"
			" FROM audio_hls_work w, bible_files bf"
			" WHERE w.hash_id = bf.hash_id"
			" AND w.file_id = bf.id")
		results = {}
		resultSet = self.select(sql, ())
		for row in resultSet:
			key = "%s:%s:%d:%d" % (row[0], row[1], row[2], row[3])
			results[key] = row[4]
		return results


    ## Returns the timestamps of a fileset in a map  fileset_id:book:chapter : [(timestamp_id, timestamp)]
	def selectTimestamps(self):
		sql = ("SELECT work.fileset_id, bf.book_id, bf.chapter_start, ft.id, ft.verse_start, ft.timestamp"
			" FROM bible_files bf, audio_hls_work work, bible_file_timestamps ft"
			" WHERE bf.hash_id = work.hash_id"
			" AND bf.id = work.file_id"
			" AND bf.id = ft.bible_file_id"
			" ORDER BY bf.file_name, ft.verse_start")
		resultSet = self.select(sql, ())
		results = {}
		for row in resultSet:
			key = "%s:%s:%d" % (row[0], row[1], row[2]) # fileset_id:book:chapter
			times = results.get(key, [])
			times.append((row[3], row[5]))
			results[key] = times
		return results


	def beginFilesetInsertTran(self):
		try:
			self.db.begin()
			self.tranCursor = self.db.cursor()
		except Exception as err:
			self.error(None, sql, err)


	def deleteSAFilesetFiles(self, hashId):
		sql = []
		sql.append("DELETE bfss FROM bible_file_stream_bytes AS bfss"
			" JOIN bible_file_stream_bandwidths AS bfsb ON bfss.stream_bandwidth_id = bfsb.id"
			" JOIN bible_files AS bf ON bfsb.bible_file_id = bf.id"
			" WHERE bf.hash_id = %s")
		sql.append("DELETE bfsb FROM bible_file_stream_bandwidths AS bfsb"
			" JOIN bible_files AS bf ON bfsb.bible_file_id = bf.id"
			" WHERE bf.hash_id = %s")
		sql.append("DELETE bft FROM bible_file_tags AS bft"
			" JOIN bible_files AS bf ON bft.file_id = bf.id"
			" WHERE bf.hash_id = %s")
		sql.append("DELETE FROM bible_files WHERE hash_id = %s")
		sql.append("DELETE FROM access_group_filesets WHERE hash_id = %s")
		sql.append("DELETE FROM bible_fileset_connections WHERE hash_id = %s")
		sql.append("DELETE FROM bible_fileset_tags WHERE hash_id = %s")
		sql.append("DELETE FROM bible_filesets WHERE hash_id = %s")
		try:
			for stmt in sql:
				self.tranCursor.execute(stmt, (hashId,))
		except Exception as err:
			self.error(self.tranCursor, sql, err)


    ## Inserts a new row into the fileset table
	def insertFileset(self, values):
		sql = ("INSERT INTO bible_filesets (hash_id, id, asset_id, set_type_code, set_size_code)"
			" VALUES (%s, %s, %s, %s, %s)")
		try:
			self.sqlLog.write((sql + "\n") % values)
			self.tranCursor.execute(sql, values)
			self.currHashId = values[0]
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def copyFilesetTables(self, saHashId, daHashId):
		values = (saHashId, daHashId,)
		sql = ""
		try:
			# pkey hash_id, name, language_id
			sql = ("INSERT INTO bible_fileset_tags (hash_id, name, description,"
				" admin_only, notes, iso, language_id)"
				" SELECT %s, name, description, admin_only, notes, iso, language_id"
				" FROM bible_fileset_tags WHERE hash_id = %s AND name != 'bitrate'")
			self.tranCursor.execute(sql, values)
			# pkey hash_id, bible_id
			sql = ("INSERT INTO bible_fileset_connections (hash_id, bible_id)"
				" SELECT %s, bible_id"
				" FROM bible_fileset_connections"
				" WHERE hash_id = %s")
			self.tranCursor.execute(sql, values)
			# pkey access_group_id, hash_id
			sql = ("INSERT INTO access_group_filesets (hash_id, access_group_id)"
				" SELECT %s, access_group_id"
				" FROM access_group_filesets"
				" WHERE hash_id = %s")
			self.tranCursor.execute(sql, values)
		except Exception as err:
			self.error(self.tranCursor, sql, err)


    ## Inserts a new row into the bible_files table
	def insertFiles(self, values):
		sql = ("INSERT INTO bible_files (hash_id, file_name, book_id, chapter_start, chapter_end,"
			" verse_start, verse_end, file_size, duration) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)")
		try:
			values = (self.currHashId,) + values
			self.sqlLog.write((sql + "\n") % values)
			self.tranCursor.execute(sql, values)
			self.currFileId = self.tranCursor.lastrowid
			self.sqlLog.write("insert bible_files, last insert id %d\n" % self.currFileId)
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def copyFileTags(self, origFileId):
		sql = ("INSERT INTO bible_file_tags (file_id, tag, value, admin_only)"
			" SELECT '%s', tag, value, admin_only"
			" FROM bible_file_tags"
			" WHERE file_id = %s AND tag != 'duration'")
		try:
			self.tranCursor.execute(sql, (self.currFileId, origFileId))
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	## Inserts a new row into the bible_file_stream_bandwidth table
	def insertBandwidth(self, values):
		sql = ("INSERT INTO bible_file_stream_bandwidths (bible_file_id, file_name, bandwidth, codec, stream)"
			" VALUES (%s, %s, %s, 'avc1.4d001f,mp4a.40.2', 1)")
		try:
			values = (self.currFileId,) + values
			self.sqlLog.write((sql + "\n") % values)
			self.tranCursor.execute(sql, values)
			self.currBandwidthId = self.tranCursor.lastrowid
			self.sqlLog.write("insert bandwidths, last_insert_id %d\n" % self.currBandwidthId)
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def addSegment(self, values):
		self.segments.append((self.currBandwidthId,) + values)


    ## Inserts a collection of rows into the bible_file_stream_bytes table
	def insertSegments(self):
		sql = ("INSERT INTO bible_file_stream_bytes (stream_bandwidth_id, timestamp_id, runtime, offset, bytes)"
			+ " VALUES (%s, %s, %s, %s, %s)")
		try:
			for segment in self.segments:
				self.sqlLog.write((sql + "\n") % segment)
			self.tranCursor.executemany(sql, self.segments)
			self.segments = []
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def copyTimestamps(self, saHashId, daHashId):
		sql = ("REPLACE INTO bible_file_timestamps (bible_file_id, verse_start, verse_end, `timestamp`)"
			" SELECT distinct bf_sa.id, t_da.verse_start, t_da.verse_end, t_da.`timestamp`"
			" FROM bible_file_timestamps t_da, bible_files bf_da, bible_filesets bs_da,"
			" bible_files bf_sa, bible_filesets bs_sa"
			" WHERE bf_da.id = t_da.bible_file_id"
			" AND bf_sa.hash_id = bs_sa.hash_id"
			" AND bf_da.hash_id = bs_da.hash_id"
			" AND bf_da.book_id = bf_sa.book_id"
			" AND bf_da.chapter_start = bf_sa.chapter_start"
			" AND bs_da.hash_id = %s"
			" AND bs_sa.hash_id = %s")
			#" AND bs_da.id = %s"
			#" AND bs_sa.id = %s")
		try:
			values = (daHashId, saHashId)
			self.sqlLog.write((sql + "\n") % values)
			self.tranCursor.execute(sql, values)
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def commitFilesetInsertTran(self):
		try:
			self.db.commit()
			self.tranCursor.close()
		except Exception as err:
			self.error(self.tranCursor, sql, err)


	def rollbackFilesetInsertTran(self):
		try:
			self.db.rollback()
			self.tranCursor.close()
		except Exception as err:
			self.error(self.tranCursor, sql, err)

##
## Convenience Methods
##
	def select(self, statement, values):
		cursor = self.db.cursor()
		try:
			cursor.execute(statement, values)
			resultSet = cursor.fetchall()
			cursor.close()
			return resultSet
		except Exception as err:
			self.error(cursor, statement, err)


	def selectScalar(self, statement, values):
		cursor = self.db.cursor()
		try:
			cursor.execute(statement, values)
			result = cursor.fetchone()
			cursor.close()
			return result[0] if result != None else None
		except Exception as err:
			self.error(cursor, statement, err)


	def selectRow(self, statement, values):
		resultSet = self.select(statement, values)
		return resultSet[0] if len(resultSet) > 0 else [None] * 10


	def selectList(self, statement, values):
		resultSet = self.select(statement, values)
		results = []
		for row in resultSet:
			results.append(row[0])
		return results


	def execute(self, statement, values):
		cursor = self.db.cursor()
		try :
			cursor.execute(statement, values)
			self.db.commit()
			cursor.close()
		except Exception as err:
			self.error(cursor, statement, err)


	def executeBatch(self, statement, valuesList):
		cursor = self.db.cursor()
		try:
			cursor.executemany(statement, valuesList)
			self.db.commit()
			cursor.close()
		except Exception as err:
			self.error(cursor, statement, err)


	def error(self, cursor, statement, err):
		if cursor != None:
			cursor.close()
		print("ERROR executing SQL %s on '%s'" % (err, statement))
		self.db.rollback()
		sys.exit()


config = Config()
hls = AudioHLS(config)
hls.processCommandLine()

