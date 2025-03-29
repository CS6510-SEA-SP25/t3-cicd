create database if not exists CicdApplication;
use CicdApplication;


drop table if exists Dependencies;
drop table if exists Jobs;
drop table if exists Stages;
drop table if exists Pipelines;

-- Pipeline's execution report
CREATE TABLE Pipelines (
	pipeline_id int auto_increment,						            -- Pipeline Execution Id								
	repository varchar(255) not null,					            -- Repository this pipeline is running on
    commit_hash varchar(255) not null default 'HEAD',	            -- Git commit hash
    ip_address varchar(255) not null default '0.0.0.0',	            -- IP address where the pipeline is executed
    
    name varchar(255) not null,							            -- Pipeline name
    stage_order varchar(1000) not null,					            -- Stage execution order
    -- exec_order?
    
    status enum('SUCCESS', 'FAILED', 'CANCELED', 'PENDING'),		-- Pipeline execution status
    start_time timestamp not null default CURRENT_TIMESTAMP,        -- Pipeline execution start time
    end_time timestamp,						            			-- Pipeline execution end time
    
    constraint pk_Pipelines_pipeline_id  primary key (pipeline_id)
);

-- Stage's execution report
CREATE TABLE Stages (
	stage_id int auto_increment,						            -- Stage Execution Id
    pipeline_id int,									            -- Pipeline Execution Id		
    
    name varchar(255) not null,							            -- Stage name		    
    
	status enum('SUCCESS', 'FAILED', 'CANCELED', 'PENDING'),		-- Pipeline execution status
    start_time timestamp not null default CURRENT_TIMESTAMP,        -- Pipeline execution start time
    end_time timestamp,						            			-- Pipeline execution end time
    
    constraint pk_Stages_stage_id primary key (stage_id),
    constraint fk_Stages_pipeline_id foreign key (pipeline_id)
		references Pipelines(pipeline_id)
        on update cascade
        on delete cascade
);

-- Job's execution report
CREATE TABLE Jobs (
	job_id int auto_increment,							            -- Job Execution Id
    stage_id int,										            -- Stage Execution Id
	
    name varchar(255) not null,							            -- Job name		
    image varchar(255) not null,						            -- Job image		
    script varchar(1000) not null,						            -- Job script		
    
    status enum('SUCCESS', 'FAILED', 'CANCELED', 'PENDING'),		-- Pipeline execution status
    start_time timestamp not null default CURRENT_TIMESTAMP,        -- Pipeline execution start time
    end_time timestamp,						            			-- Pipeline execution end time

    container_id varchar(255) not null,                             -- Docker Container ID. Used to retrieve logs.
    
    constraint pk_Jobs_job_id primary key (job_id),
    constraint fk_Jobs_stage_id foreign key (stage_id)
		references Stages(stage_id)
        on update cascade
        on delete cascade
);


-- -- Dependencies
-- CREATE TABLE Dependencies (
-- 	parent_id int,
--     child_id int,
-- 	PRIMARY KEY (parent_id, child_id),
--     FOREIGN KEY (parent_id) REFERENCES Jobs(job_id),
--     FOREIGN KEY (child_id) REFERENCES Jobs(job_id)
-- );








-- INSERT INTO Pipelines (repository, commit_hash, ip_address, name, stage_order, status, start_time)
-- VALUES 
-- ('https://github.com/example/repo', 'abc123', '192.168.1.1', 'CI Pipeline', 'Build,Test,Deploy', 'SUCCESS', '2023-10-01 10:00:00');

-- INSERT INTO Stages (pipeline_id, name, status, start_time, end_time)
-- VALUES 
-- (1, 'Build', 'SUCCESS', '2023-10-01 10:00:00', '2023-10-01 10:10:00'),
-- (1, 'Test', 'SUCCESS', '2023-10-01 10:10:00', '2023-10-01 10:20:00'),
-- (1, 'Deploy', 'SUCCESS', '2023-10-01 10:20:00', '2023-10-01 10:30:00');


-- INSERT INTO Jobs (stage_id, name, image, script, status, start_time, end_time)
-- VALUES 
-- (1, 'Compile Code', 'gcc:latest', 'gcc -o main main.c', 'SUCCESS', '2023-10-01 10:00:00', '2023-10-01 10:05:00'),
-- (1, 'Package Artifacts', 'maven:latest', 'mvn package', 'SUCCESS', '2023-10-01 10:05:00', '2023-10-01 10:10:00'),
-- (2, 'Run Unit Tests', 'node:latest', 'npm test', 'SUCCESS', '2023-10-01 10:10:00', '2023-10-01 10:15:00'),
-- (2, 'Run Integration Tests', 'python:latest', 'pytest', 'SUCCESS', '2023-10-01 10:15:00', '2023-10-01 10:20:00'),
-- (3, 'Deploy to Staging', 'kubectl:latest', 'kubectl apply -f staging.yaml', 'SUCCESS', '2023-10-01 10:20:00', '2023-10-01 10:25:00'),
-- (3, 'Deploy to Production', 'kubectl:latest', 'kubectl apply -f production.yaml', 'SUCCESS', '2023-10-01 10:25:00', '2023-10-01 10:30:00');

-- INSERT INTO Dependencies (parent_id, child_id)
-- VALUES 
-- (1, 2), -- Compile Code -> Package Artifacts
-- (2, 3), -- Package Artifacts -> Run Unit Tests
-- (3, 4), -- Run Unit Tests -> Run Integration Tests
-- (4, 5), -- Run Integration Tests -> Deploy to Staging
-- (5, 6); -- Deploy to Staging -> Deploy to Production

