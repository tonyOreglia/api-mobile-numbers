CREATE TABLE IF NOT EXISTS numbers (
  number        	text NOT NULL,
  file_ref      	uuid NOT NULL,
  country_ioc_code 	text NOT NULL,
  PRIMARY KEY (number, file_ref)
);


CREATE TABLE IF NOT EXISTS fixed_numbers (
  original_number         text NOT NULL,
  changes                 text NOT NULL,
  fixed_number            text NOT NULL,
  file_ref      		      uuid NOT NULL,
  PRIMARY KEY   		      original_number, file_ref)
);

CREATE TABLE IF NOT EXISTS rejected_numbers (
  number        text NOT NULL,
  file_ref      uuid NOT NULL,
  PRIMARY KEY   (number, file_ref)
);