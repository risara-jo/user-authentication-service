-- Smart Transit System Database Creation Script for PostgreSQL
-- Author: System Generated
-- Date: September 7, 2025
-- Version: 1.0 (PostgreSQL)

-- Create database (run this separately if needed)
-- DROP DATABASE IF EXISTS smart_transit_system;
-- CREATE DATABASE smart_transit_system;

-- Connect to the database and set up extensions
\c smart_transit_system;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom types (ENUMs)
CREATE TYPE user_role AS ENUM ('passenger', 'driver', 'conductor', 'bus_owner', 'lounge_owner', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending_verification');
CREATE TYPE verification_status AS ENUM ('pending', 'verified', 'rejected', 'suspended');
CREATE TYPE bus_type AS ENUM ('standard', 'deluxe', 'semi_sleeper', 'sleeper', 'ac', 'non_ac');
CREATE TYPE bus_status AS ENUM ('active', 'maintenance', 'inactive', 'out_of_service');
CREATE TYPE route_status AS ENUM ('active', 'inactive', 'maintenance');
CREATE TYPE staff_type AS ENUM ('driver', 'conductor');
CREATE TYPE background_check_status AS ENUM ('pending', 'cleared', 'failed');
CREATE TYPE employment_status AS ENUM ('active', 'inactive', 'suspended', 'terminated');
CREATE TYPE shift_type AS ENUM ('morning', 'evening', 'full_day', 'night');
CREATE TYPE assignment_status AS ENUM ('active', 'completed', 'cancelled');
CREATE TYPE route_type AS ENUM ('express', 'regular', 'limited_stop');
CREATE TYPE schedule_status AS ENUM ('active', 'inactive', 'cancelled');
CREATE TYPE trip_status AS ENUM ('scheduled', 'boarding', 'departed', 'in_transit', 'arrived', 'cancelled', 'delayed');
CREATE TYPE seat_type AS ENUM ('regular', 'premium', 'wheelchair', 'reserved');
CREATE TYPE seat_booking_status AS ENUM ('booked', 'cancelled', 'no_show', 'occupied');
CREATE TYPE booking_status AS ENUM ('confirmed', 'cancelled', 'completed', 'no_show');
CREATE TYPE payment_status_type AS ENUM ('pending', 'completed', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('card', 'bank_transfer', 'mobile_payment', 'cash', 'wallet');
CREATE TYPE booking_source AS ENUM ('mobile_app', 'website', 'phone', 'lounge', 'agent');
CREATE TYPE ticket_type AS ENUM ('regular', 'return', 'season_pass');
CREATE TYPE payment_status_extended AS ENUM ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded');
CREATE TYPE location_source AS ENUM ('gps', 'network', 'passive');
CREATE TYPE rating_type AS ENUM ('trip', 'driver', 'conductor', 'bus', 'lounge', 'service');
CREATE TYPE passenger_gender AS ENUM ('male', 'female', 'other');
CREATE TYPE maintenance_type AS ENUM ('routine', 'repair', 'inspection', 'emergency', 'upgrade');
CREATE TYPE maintenance_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled');
CREATE TYPE incident_type AS ENUM ('accident', 'breakdown', 'medical', 'security', 'weather', 'other');
CREATE TYPE incident_severity AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE incident_status AS ENUM ('reported', 'responding', 'resolved', 'under_investigation');
CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount', 'free_trip');
CREATE TYPE setting_type AS ENUM ('string', 'integer', 'decimal', 'boolean', 'json');
CREATE TYPE lounge_status AS ENUM ('active', 'inactive', 'maintenance', 'suspended');
CREATE TYPE lounge_booking_status AS ENUM ('confirmed', 'checked_in', 'checked_out', 'cancelled', 'no_show');
CREATE TYPE notification_type AS ENUM ('booking_confirmation', 'trip_reminder', 'bus_delay', 'bus_arrival', 'payment_success', 'lounge_booking', 'system_alert', 'route_update', 'promotion', 'maintenance_alert', 'emergency');
CREATE TYPE notification_channel AS ENUM ('push', 'sms', 'email', 'in_app');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'delivered', 'failed', 'read');
CREATE TYPE notification_priority AS ENUM ('low', 'normal', 'high', 'urgent');
CREATE TYPE device_type AS ENUM ('android', 'ios', 'web');
CREATE TYPE transaction_type AS ENUM ('credit', 'debit', 'refund', 'bonus', 'cashback');
CREATE TYPE reference_type AS ENUM ('booking', 'topup', 'refund', 'promotion', 'admin_adjustment');
CREATE TYPE wallet_transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'login', 'logout', 'payment', 'booking', 'cancellation');

-- Users table (Modified for Firebase Auth)
CREATE TABLE users (
    id VARCHAR(128) PRIMARY KEY, -- Firebase UID (not auto-generated)
    email VARCHAR(255) UNIQUE NOT NULL,
    -- password_hash removed (Firebase handles passwords)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    role user_role NOT NULL,
    status user_status DEFAULT 'active',
    email_verified BOOLEAN DEFAULT false, -- Synced from Firebase
    firebase_custom_claims JSONB, -- Store Firebase custom claims locally
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bus owners/companies
CREATE TABLE bus_owners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE, -- Firebase UID
    company_name VARCHAR(255) NOT NULL,
    license_number VARCHAR(100) UNIQUE,
    contact_person VARCHAR(255),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Sri Lanka',
    postal_code VARCHAR(20),
    verification_status verification_status DEFAULT 'pending',
    verification_documents JSONB, -- Store document URLs/metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Buses owned by companies
CREATE TABLE buses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_owner_id UUID NOT NULL,
    bus_number VARCHAR(50) NOT NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    bus_type bus_type NOT NULL,
    total_seats INTEGER NOT NULL,
    amenities JSONB, -- {"wifi": true, "ac": true, "charging_ports": true, "entertainment": false}
    manufacturing_year INTEGER,
    last_maintenance_date DATE,
    insurance_expiry DATE,
    status bus_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE
);

-- Master routes (government-defined routes)
CREATE TABLE master_routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_number VARCHAR(50) NOT NULL UNIQUE, -- e.g., "1", "138", "177" (unique in Sri Lanka)
    route_name VARCHAR(255) NOT NULL, -- e.g., "Colombo - Kandy Express"
    full_origin_city VARCHAR(100) NOT NULL, -- Starting city of complete route
    full_destination_city VARCHAR(100) NOT NULL, -- Ending city of complete route
    total_distance_km DECIMAL(8,2),
    estimated_full_duration_minutes INTEGER,
    route_description TEXT,
    status route_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- All stops for each master route (complete route stops)
CREATE TABLE master_route_stops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_route_id UUID NOT NULL,
    stop_name VARCHAR(255) NOT NULL,
    stop_order INTEGER NOT NULL, -- Sequential order in the complete route
    arrival_time_offset_minutes INTEGER, -- Minutes from route start
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    stop_description TEXT, -- Landmarks, facilities available
    is_major_stop BOOLEAN DEFAULT false, -- Major bus stands/terminals
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (master_route_id) REFERENCES master_routes(id) ON DELETE CASCADE
);

-- Bus staff (drivers and conductors)
CREATE TABLE bus_staff (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE, -- Firebase UID
    bus_owner_id UUID NOT NULL,
    staff_type staff_type NOT NULL,
    license_number VARCHAR(100),
    license_expiry_date DATE,
    experience_years INTEGER DEFAULT 0,
    emergency_contact VARCHAR(20),
    medical_certificate_expiry DATE,
    background_check_status background_check_status DEFAULT 'pending',
    employment_status employment_status DEFAULT 'active',
    hire_date DATE NOT NULL,
    termination_date DATE NULL,
    salary_amount DECIMAL(10,2),
    performance_rating DECIMAL(3,2) DEFAULT 5.00, -- Out of 5.00
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE,
    CONSTRAINT chk_performance_rating CHECK (performance_rating >= 0.00 AND performance_rating <= 5.00)
);

-- Bus staff assignments
CREATE TABLE bus_staff_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    driver_id UUID, -- Reference to bus_staff table
    conductor_id UUID, -- Reference to bus_staff table
    bus_owner_id UUID NOT NULL,
    
    -- Assignment period
    assignment_date DATE DEFAULT CURRENT_DATE,
    start_date DATE NOT NULL, -- When this assignment becomes effective
    end_date DATE, -- NULL means ongoing assignment
    
    -- Assignment details
    shift_type shift_type DEFAULT 'full_day',
    assignment_status assignment_status DEFAULT 'active',
    
    -- Notes
    notes TEXT, -- Reason for assignment, special instructions
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    FOREIGN KEY (driver_id) REFERENCES bus_staff(id) ON DELETE SET NULL,
    FOREIGN KEY (conductor_id) REFERENCES bus_staff(id) ON DELETE SET NULL,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE,
    
    -- Ensure at least one staff member is assigned
    CONSTRAINT chk_staff_assignment CHECK (driver_id IS NOT NULL OR conductor_id IS NOT NULL)
);

-- Specific routes operated by bus companies (instances of master routes)
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_route_id UUID NOT NULL,
    bus_owner_id UUID NOT NULL,
    route_name VARCHAR(255) NOT NULL, -- Operator's name for this route
    origin_stop_id UUID NOT NULL, -- Starting stop for this operator's route
    destination_stop_id UUID NOT NULL, -- Ending stop for this operator's route
    distance_km DECIMAL(8,2) NOT NULL,
    estimated_duration_minutes INTEGER NOT NULL,
    base_fare DECIMAL(10,2) NOT NULL,
    fare_per_km DECIMAL(8,2) DEFAULT 0.00,
    route_type route_type DEFAULT 'regular',
    status route_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (master_route_id) REFERENCES master_routes(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE,
    FOREIGN KEY (origin_stop_id) REFERENCES master_route_stops(id) ON DELETE RESTRICT,
    FOREIGN KEY (destination_stop_id) REFERENCES master_route_stops(id) ON DELETE RESTRICT
);

-- Stops for specific routes (subset of master route stops)
CREATE TABLE route_stops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL,
    master_route_stop_id UUID NOT NULL,
    stop_order INTEGER NOT NULL,
    arrival_time_offset_minutes INTEGER,
    departure_time_offset_minutes INTEGER,
    is_pickup_stop BOOLEAN DEFAULT true,
    is_drop_stop BOOLEAN DEFAULT true,
    fare_from_origin DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    FOREIGN KEY (master_route_stop_id) REFERENCES master_route_stops(id) ON DELETE CASCADE,
    UNIQUE (route_id, stop_order),
    UNIQUE (route_id, master_route_stop_id)
);

-- Bus schedules
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL,
    bus_id UUID NOT NULL,
    departure_time TIME NOT NULL,
    arrival_time TIME NOT NULL,
    days_of_week JSONB NOT NULL, -- [1,2,3,4,5,6,7] for days of week
    base_fare DECIMAL(10,2) NOT NULL,
    effective_from DATE NOT NULL,
    effective_to DATE,
    status schedule_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE
);

-- Individual trips
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    schedule_id UUID NOT NULL,
    route_id UUID NOT NULL,
    bus_id UUID NOT NULL,
    trip_date DATE NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    estimated_arrival_time TIMESTAMP NOT NULL,
    actual_departure_time TIMESTAMP NULL,
    actual_arrival_time TIMESTAMP NULL,
    status trip_status DEFAULT 'scheduled',
    driver_assignment_id UUID,
    conductor_assignment_id UUID,
    trip_notes TEXT, -- Special instructions or notes
    weather_conditions VARCHAR(100), -- Weather at departure
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    FOREIGN KEY (driver_assignment_id) REFERENCES bus_staff_assignments(id) ON DELETE SET NULL,
    FOREIGN KEY (conductor_assignment_id) REFERENCES bus_staff_assignments(id) ON DELETE SET NULL
);

-- Bus seats configuration
CREATE TABLE bus_seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    seat_number VARCHAR(10) NOT NULL,
    seat_type seat_type DEFAULT 'regular',
    position_row INTEGER NOT NULL,
    position_column CHAR(1) NOT NULL, -- A, B, C, D etc.
    is_window_seat BOOLEAN DEFAULT false,
    is_aisle_seat BOOLEAN DEFAULT false,
    additional_fare DECIMAL(8,2) DEFAULT 0.00,
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    UNIQUE (bus_id, seat_number)
);

-- Premium lounges at bus stations
CREATE TABLE lounges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lounge_owner_user_id VARCHAR(128) NOT NULL, -- Firebase UID
    lounge_name VARCHAR(255) NOT NULL,
    description TEXT,
    location_address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    capacity INTEGER NOT NULL,
    amenities JSONB, -- {"wifi": true, "food": true, "shower": true, "charging": true}
    operating_hours JSONB, -- {"monday": {"open": "06:00", "close": "22:00"}, ...}
    hourly_rate DECIMAL(8,2) NOT NULL,
    images JSONB, -- Array of image URLs
    contact_phone VARCHAR(20),
    status lounge_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (lounge_owner_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Lounge bookings
CREATE TABLE lounge_bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_reference VARCHAR(20) UNIQUE NOT NULL,
    lounge_id UUID NOT NULL,
    passenger_user_id VARCHAR(128) NOT NULL, -- Firebase UID
    check_in_time TIMESTAMP NOT NULL,
    check_out_time TIMESTAMP NOT NULL,
    guest_count INTEGER NOT NULL DEFAULT 1,
    total_amount DECIMAL(10,2) NOT NULL,
    booking_status lounge_booking_status DEFAULT 'confirmed',
    payment_status payment_status_type DEFAULT 'pending',
    actual_check_in TIMESTAMP NULL,
    actual_check_out TIMESTAMP NULL,
    special_requests TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (lounge_id) REFERENCES lounges(id) ON DELETE RESTRICT,
    FOREIGN KEY (passenger_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Customer bookings
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_reference VARCHAR(20) UNIQUE NOT NULL,
    passenger_user_id VARCHAR(128) NOT NULL, -- Firebase UID
    trip_id UUID NOT NULL,
    passenger_name VARCHAR(255) NOT NULL,
    passenger_phone VARCHAR(20),
    passenger_email VARCHAR(255),
    boarding_stop_id UUID NOT NULL,
    alghting_stop_id UUID NOT NULL,
    passenger_count INTEGER DEFAULT 1,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    booking_status booking_status DEFAULT 'confirmed',
    payment_status payment_status_type DEFAULT 'pending',
    payment_method payment_method DEFAULT 'card',
    booking_source booking_source DEFAULT 'mobile_app',
    ticket_type ticket_type DEFAULT 'regular',
    qr_code VARCHAR(255) UNIQUE, -- QR code for ticket verification
    booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP NULL,
    cancellation_reason TEXT,
    special_requests TEXT, -- Wheelchair access, etc.
    FOREIGN KEY (passenger_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE RESTRICT,
    FOREIGN KEY (boarding_stop_id) REFERENCES route_stops(id) ON DELETE RESTRICT,
    FOREIGN KEY (alghting_stop_id) REFERENCES route_stops(id) ON DELETE RESTRICT
);

-- Seat bookings per trip
CREATE TABLE seat_bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL,
    bus_seat_id UUID NOT NULL,
    passenger_name VARCHAR(255) NOT NULL,
    passenger_age INTEGER,
    passenger_gender passenger_gender,
    special_needs TEXT, -- Wheelchair, elderly, etc.
    status seat_booking_status DEFAULT 'booked',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_seat_id) REFERENCES bus_seats(id) ON DELETE CASCADE,
    UNIQUE (booking_id, bus_seat_id)
);

-- Payment transactions
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NULL, -- NULL for lounge bookings
    lounge_booking_id UUID NULL, -- NULL for bus bookings
    user_id VARCHAR(128) NOT NULL,
    transaction_id VARCHAR(100) UNIQUE NOT NULL,
    payment_gateway VARCHAR(50), -- stripe, paypal, razorpay, etc.
    gateway_transaction_id VARCHAR(255),
    amount DECIMAL(10,2) NOT NULL,
    currency CHAR(3) DEFAULT 'LKR',
    payment_method payment_method NOT NULL,
    payment_status payment_status_extended DEFAULT 'pending',
    failure_reason TEXT,
    gateway_response JSONB, -- Store gateway response
    refund_amount DECIMAL(10,2) DEFAULT 0.00,
    refund_reason TEXT,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE SET NULL,
    FOREIGN KEY (lounge_booking_id) REFERENCES lounge_bookings(id) ON DELETE SET NULL,
    CONSTRAINT chk_booking_type CHECK ((booking_id IS NOT NULL AND lounge_booking_id IS NULL) OR (booking_id IS NULL AND lounge_booking_id IS NOT NULL))
);

-- Real-time bus tracking
CREATE TABLE bus_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NOT NULL,
    bus_id UUID NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    speed_kmh DECIMAL(5,2) DEFAULT 0.00,
    heading INTEGER, -- Direction in degrees (0-359)
    altitude_meters DECIMAL(8,2),
    accuracy_meters DECIMAL(6,2),
    battery_level INTEGER, -- Tracker device battery
    is_engine_on BOOLEAN DEFAULT true,
    odometer_reading DECIMAL(10,2), -- Total distance traveled
    fuel_level_percentage INTEGER,
    location_source location_source DEFAULT 'gps',
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE
);

-- Customer feedback and ratings
CREATE TABLE feedback_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NULL,
    lounge_booking_id UUID NULL,
    user_id VARCHAR(128) NOT NULL,
    trip_id UUID NULL,
    bus_id UUID NULL,
    driver_id UUID NULL,
    conductor_id UUID NULL,
    lounge_id UUID NULL,
    rating_type rating_type NOT NULL,
    overall_rating INTEGER NOT NULL CHECK (overall_rating >= 1 AND overall_rating <= 5),
    punctuality_rating INTEGER CHECK (punctuality_rating >= 1 AND punctuality_rating <= 5),
    cleanliness_rating INTEGER CHECK (cleanliness_rating >= 1 AND cleanliness_rating <= 5),
    comfort_rating INTEGER CHECK (comfort_rating >= 1 AND comfort_rating <= 5),
    service_rating INTEGER CHECK (service_rating >= 1 AND service_rating <= 5),
    feedback_text TEXT,
    is_anonymous BOOLEAN DEFAULT false,
    is_verified_trip BOOLEAN DEFAULT false,
    response_from_operator TEXT, -- Response from bus company
    responded_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE SET NULL,
    FOREIGN KEY (lounge_booking_id) REFERENCES lounge_bookings(id) ON DELETE SET NULL,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE SET NULL,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE SET NULL,
    FOREIGN KEY (driver_id) REFERENCES bus_staff(id) ON DELETE SET NULL,
    FOREIGN KEY (conductor_id) REFERENCES bus_staff(id) ON DELETE SET NULL,
    FOREIGN KEY (lounge_id) REFERENCES lounges(id) ON DELETE SET NULL
);

-- Bus maintenance records
CREATE TABLE maintenance_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    maintenance_type maintenance_type NOT NULL,
    description TEXT NOT NULL,
    maintenance_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    cost DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    service_provider VARCHAR(255), -- Workshop/mechanic name
    parts_replaced JSONB, -- {"brake_pads": 4, "oil_filter": 1}
    next_service_due_date DATE,
    next_service_due_km INTEGER,
    performed_by VARCHAR(255), -- Technician name
    maintenance_status maintenance_status DEFAULT 'scheduled',
    notes TEXT,
    receipt_images JSONB, -- Array of receipt image URLs
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE
);

-- Emergency contacts and incidents
CREATE TABLE emergency_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NULL,
    bus_id UUID NOT NULL,
    reporter_user_id VARCHAR(128) NULL,
    incident_type incident_type NOT NULL,
    severity incident_severity NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location_latitude DECIMAL(10, 8),
    location_longitude DECIMAL(11, 8),
    location_description TEXT,
    emergency_services_called BOOLEAN DEFAULT false,
    authorities_notified BOOLEAN DEFAULT false,
    passengers_affected INTEGER DEFAULT 0,
    injuries_reported BOOLEAN DEFAULT false,
    incident_status incident_status DEFAULT 'reported',
    resolution_notes TEXT,
    incident_images JSONB, -- Array of incident image URLs
    reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE SET NULL,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    FOREIGN KEY (reporter_user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Promotional offers and discounts
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promotion_code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type discount_type NOT NULL,
    discount_value DECIMAL(8,2) NOT NULL,
    min_booking_amount DECIMAL(8,2) DEFAULT 0.00,
    max_discount_amount DECIMAL(8,2),
    usage_limit INTEGER, -- Total usage limit
    usage_per_user INTEGER DEFAULT 1,
    applicable_routes JSONB, -- Array of route IDs, null for all routes
    user_eligibility JSONB, -- {"new_users": true, "returning_users": false}
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    terms_conditions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Promotion usage tracking
CREATE TABLE promotion_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promotion_id UUID NOT NULL,
    user_id VARCHAR(128) NOT NULL,
    booking_id UUID NOT NULL,
    discount_applied DECIMAL(8,2) NOT NULL,
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (promotion_id) REFERENCES promotions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE
);

-- System configuration
CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    setting_type setting_type DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT false, -- Whether to expose to mobile app
    category VARCHAR(50) DEFAULT 'general',
    updated_by VARCHAR(128), -- Admin user ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Supported cities and regions
CREATE TABLE supported_cities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    city_name VARCHAR(100) NOT NULL,
    state_province VARCHAR(100),
    country VARCHAR(100) NOT NULL DEFAULT 'Sri Lanka',
    city_code VARCHAR(10) UNIQUE, -- Short code like CMB, KDY
    timezone VARCHAR(50) DEFAULT 'Asia/Colombo',
    currency CHAR(3) DEFAULT 'LKR',
    is_active BOOLEAN DEFAULT true,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    service_radius_km INTEGER DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- System audit logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NULL, -- NULL for system actions
    action_type audit_action NOT NULL,
    table_name VARCHAR(100),
    record_id VARCHAR(36),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    session_id VARCHAR(255),
    additional_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- System notifications (Firebase Cloud Messaging integrated)
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL, -- Firebase UID
    notification_type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB, -- Additional payload data
    channel notification_channel NOT NULL,
    status notification_status DEFAULT 'pending',
    firebase_message_id VARCHAR(255), -- FCM message ID for tracking
    scheduled_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    retry_count INTEGER DEFAULT 0,
    priority notification_priority DEFAULT 'normal',
    expires_at TIMESTAMP NULL, -- When notification becomes irrelevant
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User sessions for mobile app
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    device_id VARCHAR(255),
    device_type device_type NOT NULL,
    device_model VARCHAR(100),
    app_version VARCHAR(20),
    os_version VARCHAR(20),
    fcm_token VARCHAR(255), -- For push notifications
    ip_address VARCHAR(45),
    location_permission BOOLEAN DEFAULT false,
    notification_permission BOOLEAN DEFAULT false,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Wallet system for digital payments
CREATE TABLE user_wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE,
    balance DECIMAL(12,2) DEFAULT 0.00,
    currency CHAR(3) DEFAULT 'LKR',
    is_active BOOLEAN DEFAULT true,
    daily_limit DECIMAL(10,2) DEFAULT 10000.00,
    monthly_limit DECIMAL(12,2) DEFAULT 100000.00,
    pin_hash VARCHAR(255), -- Hashed wallet PIN
    last_transaction_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Wallet transactions
CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL,
    transaction_type transaction_type NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    balance_before DECIMAL(12,2) NOT NULL,
    balance_after DECIMAL(12,2) NOT NULL,
    description VARCHAR(255) NOT NULL,
    reference_type reference_type NOT NULL,
    reference_id UUID, -- booking_id, payment_id, etc.
    gateway_transaction_id VARCHAR(255),
    status wallet_transaction_status DEFAULT 'completed',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES user_wallets(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_buses_owner ON buses(bus_owner_id);
CREATE INDEX idx_buses_status ON buses(status);
CREATE INDEX idx_routes_master ON routes(master_route_id);
CREATE INDEX idx_routes_owner ON routes(bus_owner_id);
CREATE INDEX idx_schedules_route ON schedules(route_id);
CREATE INDEX idx_schedules_bus ON schedules(bus_id);
CREATE INDEX idx_trips_date ON trips(trip_date);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_bookings_user ON bookings(passenger_user_id);
CREATE INDEX idx_bookings_trip ON bookings(trip_id);
CREATE INDEX idx_bookings_status ON bookings(booking_status);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(payment_status);
CREATE INDEX idx_bus_locations_trip_recorded ON bus_locations(trip_id, recorded_at);
CREATE INDEX idx_bus_locations_bus_recorded ON bus_locations(bus_id, recorded_at);
CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action_type, created_at);
CREATE INDEX idx_audit_logs_table_record ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_notifications_user_status ON notifications(user_id, status);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_at);
CREATE INDEX idx_notifications_type_priority ON notifications(notification_type, priority);
CREATE INDEX idx_user_sessions_user_active ON user_sessions(user_id, is_active);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_wallet_transactions_wallet_created ON wallet_transactions(wallet_id, created_at);
CREATE INDEX idx_wallet_transactions_reference ON wallet_transactions(reference_type, reference_id);

-- Create trigger function to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bus_owners_updated_at BEFORE UPDATE ON bus_owners FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_buses_updated_at BEFORE UPDATE ON buses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_master_routes_updated_at BEFORE UPDATE ON master_routes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bus_staff_updated_at BEFORE UPDATE ON bus_staff FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bus_staff_assignments_updated_at BEFORE UPDATE ON bus_staff_assignments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_routes_updated_at BEFORE UPDATE ON routes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lounges_updated_at BEFORE UPDATE ON lounges FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_maintenance_records_updated_at BEFORE UPDATE ON maintenance_records FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_promotions_updated_at BEFORE UPDATE ON promotions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE ON system_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_sessions_updated_at BEFORE UPDATE ON user_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_wallets_updated_at BEFORE UPDATE ON user_wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some basic system settings
INSERT INTO system_settings (setting_key, setting_value, setting_type, description, is_public, category) VALUES
('app_name', 'Smart Transit System', 'string', 'Application name', true, 'general'),
('default_currency', 'LKR', 'string', 'Default currency for the system', true, 'financial'),
('max_booking_days_advance', '30', 'integer', 'Maximum days in advance for booking', true, 'booking'),
('cancellation_cutoff_hours', '2', 'integer', 'Hours before departure when cancellation is not allowed', true, 'booking'),
('default_timezone', 'Asia/Colombo', 'string', 'Default timezone for the system', true, 'general'),
('maintenance_mode', 'false', 'boolean', 'System maintenance mode flag', false, 'system'),
('api_version', '1.0', 'string', 'Current API version', true, 'system'),
('support_email', 'support@smarttransit.lk', 'string', 'Support email address', true, 'contact'),
('support_phone', '+94112345678', 'string', 'Support phone number', true, 'contact'),
('wallet_daily_limit', '10000.00', 'decimal', 'Default daily wallet limit', false, 'financial');

-- Insert some basic cities
INSERT INTO supported_cities (city_name, state_province, country, city_code, latitude, longitude) VALUES
('Colombo', 'Western', 'Sri Lanka', 'CMB', 6.9271, 79.8612),
('Kandy', 'Central', 'Sri Lanka', 'KDY', 7.2906, 80.6337),
('Galle', 'Southern', 'Sri Lanka', 'GAL', 6.0535, 80.2210),
('Jaffna', 'Northern', 'Sri Lanka', 'JAF', 9.6615, 80.0255),
('Anuradhapura', 'North Central', 'Sri Lanka', 'ANU', 8.3114, 80.4037);

-- Success message
SELECT 'Smart Transit System PostgreSQL Database Created Successfully!' as message;
