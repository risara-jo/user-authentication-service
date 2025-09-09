-- Smart Transit System Database Migration - Complete Schema
-- Author: System Generated
-- Date: September 7, 2025
-- Version: 1.0 (PostgreSQL)
-- This file creates the complete Smart Transit System database schema

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
    model VARCHAR(100),
    manufacturer VARCHAR(100),
    year_manufactured INTEGER,
    capacity INTEGER NOT NULL,
    bus_type bus_type NOT NULL,
    amenities JSONB, -- AC, WiFi, USB charging, etc.
    status bus_status DEFAULT 'active',
    last_maintenance_date DATE,
    next_maintenance_due DATE,
    insurance_expiry DATE,
    registration_expiry DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE
);

-- Staff (drivers, conductors)
CREATE TABLE staff (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE, -- Firebase UID
    bus_owner_id UUID NOT NULL,
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    staff_type staff_type NOT NULL,
    license_number VARCHAR(100), -- For drivers
    license_expiry DATE, -- For drivers
    emergency_contact VARCHAR(255),
    address TEXT,
    background_check_status background_check_status DEFAULT 'pending',
    employment_status employment_status DEFAULT 'active',
    hire_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (bus_owner_id) REFERENCES bus_owners(id) ON DELETE CASCADE
);

-- Staff shift assignments
CREATE TABLE staff_shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    staff_id UUID NOT NULL,
    shift_type shift_type NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    days_of_week INTEGER[], -- Array: 0=Sunday, 1=Monday, etc.
    status assignment_status DEFAULT 'active',
    effective_from DATE NOT NULL,
    effective_until DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(id) ON DELETE CASCADE
);

-- Routes
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_name VARCHAR(255) NOT NULL,
    route_number VARCHAR(50) UNIQUE NOT NULL,
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    distance_km DECIMAL(8,2),
    estimated_duration_minutes INTEGER,
    route_type route_type DEFAULT 'regular',
    status route_status DEFAULT 'active',
    waypoints JSONB, -- Store GPS coordinates of stops
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Route stops
CREATE TABLE route_stops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL,
    stop_name VARCHAR(255) NOT NULL,
    stop_order INTEGER NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    estimated_arrival_offset_minutes INTEGER, -- Minutes from route start
    is_major_stop BOOLEAN DEFAULT false,
    amenities JSONB, -- Shelter, seating, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    UNIQUE(route_id, stop_order)
);

-- Bus schedules
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    route_id UUID NOT NULL,
    driver_id UUID,
    conductor_id UUID,
    departure_time TIME NOT NULL,
    arrival_time TIME,
    days_of_week INTEGER[], -- Array: 0=Sunday, 1=Monday, etc.
    status schedule_status DEFAULT 'active',
    effective_from DATE NOT NULL,
    effective_until DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
    FOREIGN KEY (driver_id) REFERENCES staff(id),
    FOREIGN KEY (conductor_id) REFERENCES staff(id)
);

-- Daily trips (instances of schedules)
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    schedule_id UUID NOT NULL,
    trip_date DATE NOT NULL,
    actual_departure_time TIMESTAMP,
    actual_arrival_time TIMESTAMP,
    estimated_departure_time TIMESTAMP NOT NULL,
    estimated_arrival_time TIMESTAMP,
    status trip_status DEFAULT 'scheduled',
    driver_id UUID,
    conductor_id UUID,
    current_latitude DECIMAL(10,8),
    current_longitude DECIMAL(11,8),
    last_location_update TIMESTAMP,
    delay_minutes INTEGER DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (driver_id) REFERENCES staff(id),
    FOREIGN KEY (conductor_id) REFERENCES staff(id),
    UNIQUE(schedule_id, trip_date)
);

-- Bus seats
CREATE TABLE bus_seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    seat_number VARCHAR(10) NOT NULL,
    seat_type seat_type DEFAULT 'regular',
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE,
    UNIQUE(bus_id, seat_number)
);

-- Passenger profiles
CREATE TABLE passenger_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE, -- Firebase UID
    date_of_birth DATE,
    gender passenger_gender,
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(20),
    preferences JSONB, -- Window seat, aisle, etc.
    frequent_routes UUID[], -- Array of route IDs
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Bookings
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_reference VARCHAR(20) UNIQUE NOT NULL,
    passenger_id VARCHAR(128) NOT NULL, -- Firebase UID
    trip_id UUID NOT NULL,
    origin_stop_id UUID NOT NULL,
    destination_stop_id UUID NOT NULL,
    number_of_passengers INTEGER NOT NULL DEFAULT 1,
    ticket_type ticket_type DEFAULT 'regular',
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    final_amount DECIMAL(10,2) NOT NULL,
    status booking_status DEFAULT 'confirmed',
    booking_source booking_source DEFAULT 'mobile_app',
    payment_method payment_method,
    payment_status payment_status_type DEFAULT 'pending',
    booking_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    travel_date DATE NOT NULL,
    special_requirements TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (passenger_id) REFERENCES users(id),
    FOREIGN KEY (trip_id) REFERENCES trips(id),
    FOREIGN KEY (origin_stop_id) REFERENCES route_stops(id),
    FOREIGN KEY (destination_stop_id) REFERENCES route_stops(id)
);

-- Seat bookings (for reserved seating buses)
CREATE TABLE seat_bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL,
    seat_id UUID NOT NULL,
    passenger_name VARCHAR(255), -- For named bookings
    status seat_booking_status DEFAULT 'booked',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    FOREIGN KEY (seat_id) REFERENCES bus_seats(id),
    UNIQUE(seat_id, booking_id)
);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method payment_method NOT NULL,
    payment_status payment_status_extended DEFAULT 'pending',
    transaction_id VARCHAR(255), -- External payment gateway transaction ID
    gateway_response JSONB, -- Store payment gateway response
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE
);

-- Lounges/waiting areas
CREATE TABLE lounges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id VARCHAR(128) NOT NULL, -- Firebase UID of lounge owner
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    amenities JSONB, -- WiFi, food, seating, etc.
    capacity INTEGER,
    hourly_rate DECIMAL(8,2),
    status lounge_status DEFAULT 'active',
    operating_hours JSONB, -- Store operating hours for each day
    contact_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Lounge bookings
CREATE TABLE lounge_bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lounge_id UUID NOT NULL,
    passenger_id VARCHAR(128) NOT NULL, -- Firebase UID
    check_in_time TIMESTAMP,
    check_out_time TIMESTAMP,
    planned_duration_hours INTEGER,
    total_amount DECIMAL(8,2),
    status lounge_booking_status DEFAULT 'confirmed',
    special_requests TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (lounge_id) REFERENCES lounges(id) ON DELETE CASCADE,
    FOREIGN KEY (passenger_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Real-time bus tracking
CREATE TABLE bus_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    speed_kmh DECIMAL(5,2),
    heading INTEGER, -- Direction in degrees (0-359)
    altitude_meters INTEGER,
    accuracy_meters INTEGER,
    location_source location_source DEFAULT 'gps',
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE
);

-- Ratings and reviews
CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL, -- Firebase UID
    rating_type rating_type NOT NULL,
    target_id UUID NOT NULL, -- ID of trip, driver, conductor, bus, or lounge
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL, -- Firebase UID
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type notification_type NOT NULL,
    channel notification_channel DEFAULT 'push',
    status notification_status DEFAULT 'pending',
    priority notification_priority DEFAULT 'normal',
    data JSONB, -- Additional data for the notification
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Device tokens for push notifications
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL, -- Firebase UID
    device_type device_type NOT NULL,
    token VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, device_type, token)
);

-- Digital wallet
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL UNIQUE, -- Firebase UID
    balance DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'LKR',
    is_active BOOLEAN DEFAULT true,
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
    balance_before DECIMAL(10,2) NOT NULL,
    balance_after DECIMAL(10,2) NOT NULL,
    reference_type reference_type,
    reference_id UUID, -- Booking ID, etc.
    description TEXT,
    status wallet_transaction_status DEFAULT 'completed',
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

-- Promotions and discounts
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type discount_type NOT NULL,
    discount_value DECIMAL(8,2) NOT NULL,
    min_booking_amount DECIMAL(8,2) DEFAULT 0,
    max_discount_amount DECIMAL(8,2),
    usage_limit INTEGER,
    usage_count INTEGER DEFAULT 0,
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    applicable_routes UUID[], -- Array of route IDs (empty = all routes)
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User promotion usage tracking
CREATE TABLE user_promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128) NOT NULL, -- Firebase UID
    promotion_id UUID NOT NULL,
    booking_id UUID NOT NULL,
    discount_applied DECIMAL(8,2) NOT NULL,
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (promotion_id) REFERENCES promotions(id) ON DELETE CASCADE,
    FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    UNIQUE(user_id, promotion_id, booking_id)
);

-- Bus maintenance records
CREATE TABLE maintenance_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id UUID NOT NULL,
    maintenance_type maintenance_type NOT NULL,
    description TEXT NOT NULL,
    scheduled_date DATE,
    actual_date DATE,
    cost DECIMAL(10,2),
    service_provider VARCHAR(255),
    status maintenance_status DEFAULT 'scheduled',
    next_maintenance_date DATE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bus_id) REFERENCES buses(id) ON DELETE CASCADE
);

-- Incident reports
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID,
    bus_id UUID,
    reporter_id VARCHAR(128), -- Firebase UID
    incident_type incident_type NOT NULL,
    severity incident_severity DEFAULT 'medium',
    status incident_status DEFAULT 'reported',
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    location_latitude DECIMAL(10,8),
    location_longitude DECIMAL(11,8),
    location_description TEXT,
    incident_time TIMESTAMP NOT NULL,
    images JSONB, -- Array of image URLs
    resolution TEXT,
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(128), -- Firebase UID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trip_id) REFERENCES trips(id),
    FOREIGN KEY (bus_id) REFERENCES buses(id),
    FOREIGN KEY (reporter_id) REFERENCES users(id),
    FOREIGN KEY (resolved_by) REFERENCES users(id)
);

-- System settings/configuration
CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    setting_type setting_type DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT false, -- Whether setting is visible to non-admin users
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit log for important actions
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(128), -- Firebase UID (can be null for system actions)
    action audit_action NOT NULL,
    table_name VARCHAR(100),
    record_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create indexes for better performance

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Bus owners indexes
CREATE INDEX idx_bus_owners_user_id ON bus_owners(user_id);
CREATE INDEX idx_bus_owners_verification_status ON bus_owners(verification_status);

-- Buses indexes
CREATE INDEX idx_buses_owner_id ON buses(bus_owner_id);
CREATE INDEX idx_buses_license_plate ON buses(license_plate);
CREATE INDEX idx_buses_status ON buses(status);

-- Staff indexes
CREATE INDEX idx_staff_user_id ON staff(user_id);
CREATE INDEX idx_staff_bus_owner_id ON staff(bus_owner_id);
CREATE INDEX idx_staff_type ON staff(staff_type);
CREATE INDEX idx_staff_employment_status ON staff(employment_status);

-- Routes indexes
CREATE INDEX idx_routes_route_number ON routes(route_number);
CREATE INDEX idx_routes_status ON routes(status);

-- Route stops indexes
CREATE INDEX idx_route_stops_route_id ON route_stops(route_id);
CREATE INDEX idx_route_stops_location ON route_stops(latitude, longitude);

-- Schedules indexes
CREATE INDEX idx_schedules_bus_id ON schedules(bus_id);
CREATE INDEX idx_schedules_route_id ON schedules(route_id);
CREATE INDEX idx_schedules_status ON schedules(status);

-- Trips indexes
CREATE INDEX idx_trips_schedule_id ON trips(schedule_id);
CREATE INDEX idx_trips_date ON trips(trip_date);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_date_status ON trips(trip_date, status);

-- Bookings indexes
CREATE INDEX idx_bookings_passenger_id ON bookings(passenger_id);
CREATE INDEX idx_bookings_trip_id ON bookings(trip_id);
CREATE INDEX idx_bookings_reference ON bookings(booking_reference);
CREATE INDEX idx_bookings_travel_date ON bookings(travel_date);
CREATE INDEX idx_bookings_status ON bookings(status);

-- Bus tracking indexes
CREATE INDEX idx_bus_tracking_trip_id ON bus_tracking(trip_id);
CREATE INDEX idx_bus_tracking_timestamp ON bus_tracking(timestamp);

-- Notifications indexes
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_type ON notifications(type);

-- Wallet transactions indexes
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_date ON wallet_transactions(transaction_date);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_table_record ON audit_logs(table_name, record_id);

-- Insert some default system settings
INSERT INTO system_settings (setting_key, setting_value, setting_type, description, is_public) VALUES
('app_version', '1.0.0', 'string', 'Current application version', true),
('maintenance_mode', 'false', 'boolean', 'Whether the system is in maintenance mode', true),
('booking_advance_days', '30', 'integer', 'Maximum days in advance for booking', true),
('cancellation_hours', '24', 'integer', 'Hours before trip for free cancellation', true),
('default_currency', 'LKR', 'string', 'Default system currency', true),
('max_passengers_per_booking', '6', 'integer', 'Maximum passengers per booking', true),
('rating_required_for_rebooking', 'false', 'boolean', 'Whether rating is required before rebooking', false);

-- Success message
SELECT 'Smart Transit System Database Created Successfully!' as message;
