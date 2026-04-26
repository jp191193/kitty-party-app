-- =============================================
-- Script Name   : seed_member_profiles.sql
-- Description   : Dummy profile data for all existing members
-- Usage         : psql -d kitty_party_db -f db/seeds/seed_member_profiles.sql
-- =============================================

BEGIN;

INSERT INTO member_profiles
    (member_id, date_of_birth, profile_picture_url, address, city, state, pincode, occupation, about)
VALUES
    (
        '8093a455-df19-40ff-9602-60650338ed26',  -- Anjali Sharma
        '1990-05-15',
        'https://ui-avatars.com/api/?name=Anjali+Sharma&background=f06292&color=fff',
        'B-12, Shivaji Nagar, Near SBI Bank',
        'Mumbai',
        'Maharashtra',
        '400001',
        'School Teacher',
        'Primary school teacher with 10 years of experience. Loves kitty parties and saving together.'
    ),
    (
        '3011a091-f458-4801-be3e-6ba065f798e6',  -- Priya Mehta
        '1988-11-22',
        'https://ui-avatars.com/api/?name=Priya+Mehta&background=ce93d8&color=fff',
        'Flat 304, Oberoi Heights, Andheri East',
        'Mumbai',
        'Maharashtra',
        '400069',
        'Software Engineer',
        'Works at a tech startup. Believes in disciplined savings through rotating chit funds.'
    ),
    (
        '1b48b4b1-6d0b-4522-83be-d6f35b9fa083',  -- Rahul Desai
        '1985-03-10',
        'https://ui-avatars.com/api/?name=Rahul+Desai&background=80cbc4&color=fff',
        '22, Koregaon Park, Lane 7',
        'Pune',
        'Maharashtra',
        '411001',
        'Business Owner',
        'Runs a textile business in Pune. Participates in kitty parties to build business capital.'
    ),
    (
        '8b1b856f-b5e7-4e14-8378-4447266871fb',  -- Neha Singh
        '1992-07-08',
        'https://ui-avatars.com/api/?name=Neha+Singh&background=ef9a9a&color=fff',
        'A-403, Defence Colony, Block C',
        'New Delhi',
        'Delhi',
        '110024',
        'Doctor',
        'MBBS, MD (Paediatrics). Joins kitty parties with colleagues to fund family vacations.'
    ),
    (
        'a242e558-2f30-45a7-8305-94f5298cacf5',  -- Vikram Patel
        '1983-12-30',
        'https://ui-avatars.com/api/?name=Vikram+Patel&background=a5d6a7&color=fff',
        '9, Satellite Road, Jodhpur Village',
        'Ahmedabad',
        'Gujarat',
        '380015',
        'Chartered Accountant',
        'CA with 15 years of practice. Uses kitty savings to invest in fixed deposits.'
    ),
    (
        'a3b88af7-25f9-4a16-a09c-555a7d073a64',  -- Priya Kumari
        '1995-09-14',
        'https://ui-avatars.com/api/?name=Priya+Kumari&background=ffcc80&color=fff',
        '56, Rajendra Nagar, Near Patna Junction',
        'Patna',
        'Bihar',
        '800001',
        'HR Manager',
        'People-first HR professional. Organises office kitty parties to boost team bonding.'
    ),
    (
        'c1e663a4-50df-435b-b8c4-b5eea7369ebf',  -- Vandana Ponkia
        '1991-02-28',
        'https://ui-avatars.com/api/?name=Vandana+Ponkia&background=90caf9&color=fff',
        '101, Diamond Residency, Vesu',
        'Surat',
        'Gujarat',
        '395007',
        'Entrepreneur',
        'Runs a boutique clothing brand. Uses kitty funds for seasonal inventory buying.'
    ),
    (
        '172fcbef-60ec-49f6-a2b7-9e01c05875a2',  -- Swapnil Singh
        '1987-06-17',
        'https://ui-avatars.com/api/?name=Swapnil+Singh&background=b0bec5&color=fff',
        'H-22, Sector 62, Near Metro Station',
        'Noida',
        'Uttar Pradesh',
        '201301',
        'Data Analyst',
        'Data analyst at an MNC. Participates in kitty parties to save for home loan down payment.'
    ),
    (
        '50a007eb-cdff-43cb-8b41-49f613dab1ef',  -- Namrata Bulbule
        '1993-04-05',
        'https://ui-avatars.com/api/?name=Namrata+Bulbule&background=f48fb1&color=fff',
        '7, Gangapur Road, Nashik West',
        'Nashik',
        'Maharashtra',
        '422005',
        'Marketing Manager',
        'Digital marketing enthusiast. Saves through kitty parties to fund her travel bucket list.'
    )
ON CONFLICT (member_id) DO NOTHING;

COMMIT;
