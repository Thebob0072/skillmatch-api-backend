--
-- PostgreSQL database dump
--

\restrict QAjHrWmPLXykyZ4sBU13UNRfv1TH2cCBma9EeE8peTrynIw2Tm5pECq6iSD8Q18

-- Dumped from database version 15.14
-- Dumped by pg_dump version 15.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: calculate_provider_earning(numeric); Type: FUNCTION; Schema: public; Owner: admin
--

CREATE FUNCTION public.calculate_provider_earning(booking_amount numeric) RETURNS TABLE(gross_amount numeric, stripe_fee numeric, platform_commission numeric, total_fee numeric, net_amount numeric, provider_percentage numeric)
    LANGUAGE plpgsql
    AS $$
		BEGIN
			RETURN QUERY
			SELECT 
				booking_amount,
				ROUND(booking_amount * 0.0275, 2),
				ROUND(booking_amount * 0.1000, 2),
				ROUND(booking_amount * 0.1275, 2),
				ROUND(booking_amount * 0.8725, 2),
				87.25;
		END;
		$$;


ALTER FUNCTION public.calculate_provider_earning(booking_amount numeric) OWNER TO admin;

--
-- Name: update_schedule_timestamp(); Type: FUNCTION; Schema: public; Owner: admin
--

CREATE FUNCTION public.update_schedule_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$;


ALTER FUNCTION public.update_schedule_timestamp() OWNER TO admin;

--
-- Name: update_user_face_verification(); Type: FUNCTION; Schema: public; Owner: admin
--

CREATE FUNCTION public.update_user_face_verification() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- เมื่อ face verification ถูก approve ให้อัพเดท user
    IF NEW.verification_status = 'approved' AND OLD.verification_status != 'approved' THEN
        UPDATE users 
        SET face_verified = true,
            face_verification_id = NEW.verification_id
        WHERE user_id = NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_user_face_verification() OWNER TO admin;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: bank_accounts; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.bank_accounts (
    bank_account_id integer NOT NULL,
    user_id integer NOT NULL,
    bank_name character varying(100) NOT NULL,
    account_number character varying(50) NOT NULL,
    account_holder_name character varying(255) NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.bank_accounts OWNER TO admin;

--
-- Name: bank_accounts_bank_account_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.bank_accounts_bank_account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.bank_accounts_bank_account_id_seq OWNER TO admin;

--
-- Name: bank_accounts_bank_account_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.bank_accounts_bank_account_id_seq OWNED BY public.bank_accounts.bank_account_id;


--
-- Name: blocks; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.blocks (
    block_id integer NOT NULL,
    blocker_id integer NOT NULL,
    blocked_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.blocks OWNER TO admin;

--
-- Name: blocks_block_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.blocks_block_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.blocks_block_id_seq OWNER TO admin;

--
-- Name: blocks_block_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.blocks_block_id_seq OWNED BY public.blocks.block_id;


--
-- Name: bookings; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.bookings (
    booking_id integer NOT NULL,
    client_id integer NOT NULL,
    provider_id integer NOT NULL,
    package_id integer NOT NULL,
    booking_date date NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    total_price numeric(10,2) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    location text,
    special_notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    cancellation_reason text
);


ALTER TABLE public.bookings OWNER TO admin;

--
-- Name: bookings_booking_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.bookings_booking_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.bookings_booking_id_seq OWNER TO admin;

--
-- Name: bookings_booking_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.bookings_booking_id_seq OWNED BY public.bookings.booking_id;


--
-- Name: commission_transactions; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.commission_transactions (
    commission_txn_id integer NOT NULL,
    booking_id integer,
    transaction_id integer,
    booking_amount numeric(12,2) NOT NULL,
    commission_rate numeric(5,4) DEFAULT 0.1000,
    commission_amount numeric(12,2) NOT NULL,
    provider_amount numeric(12,2) NOT NULL,
    provider_id integer NOT NULL,
    platform_bank_account_id integer,
    status character varying(20) DEFAULT 'collected'::character varying,
    collected_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    refunded_at timestamp without time zone,
    refund_reason text,
    notes text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.commission_transactions OWNER TO admin;

--
-- Name: commission_transactions_commission_txn_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.commission_transactions_commission_txn_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.commission_transactions_commission_txn_id_seq OWNER TO admin;

--
-- Name: commission_transactions_commission_txn_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.commission_transactions_commission_txn_id_seq OWNED BY public.commission_transactions.commission_txn_id;


--
-- Name: conversations; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.conversations (
    conversation_id integer NOT NULL,
    user1_id integer NOT NULL,
    user2_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT user_order CHECK ((user1_id < user2_id))
);


ALTER TABLE public.conversations OWNER TO admin;

--
-- Name: conversations_conversation_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.conversations_conversation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.conversations_conversation_id_seq OWNER TO admin;

--
-- Name: conversations_conversation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.conversations_conversation_id_seq OWNED BY public.conversations.conversation_id;


--
-- Name: face_verifications; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.face_verifications (
    verification_id integer NOT NULL,
    user_id integer NOT NULL,
    selfie_url text NOT NULL,
    liveness_video_url text,
    match_confidence numeric(5,2),
    is_match boolean DEFAULT false,
    national_id_photo_url text,
    liveness_passed boolean DEFAULT false,
    liveness_confidence numeric(5,2),
    verification_status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    api_provider character varying(50),
    api_response_data jsonb,
    created_at timestamp with time zone DEFAULT now(),
    verified_at timestamp with time zone,
    verified_by integer,
    rejection_reason text,
    retry_count integer DEFAULT 0,
    document_type character varying(20) DEFAULT 'national_id'::character varying NOT NULL,
    document_id integer,
    CONSTRAINT face_verifications_document_type_check CHECK (((document_type)::text = ANY ((ARRAY['national_id'::character varying, 'passport'::character varying])::text[])))
);


ALTER TABLE public.face_verifications OWNER TO admin;

--
-- Name: TABLE face_verifications; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON TABLE public.face_verifications IS 'ระบบยืนยันใบหน้า Provider (Face Recognition + Liveness Detection)';


--
-- Name: COLUMN face_verifications.match_confidence; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON COLUMN public.face_verifications.match_confidence IS 'ความแม่นยำในการจับคู่ใบหน้ากับบัตรประชาชน (0-100%)';


--
-- Name: COLUMN face_verifications.liveness_passed; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON COLUMN public.face_verifications.liveness_passed IS 'ผ่าน liveness detection (ป้องกันการใช้รูปถ่าย)';


--
-- Name: COLUMN face_verifications.api_response_data; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON COLUMN public.face_verifications.api_response_data IS 'เก็บ raw response จาก face recognition API';


--
-- Name: COLUMN face_verifications.document_type; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON COLUMN public.face_verifications.document_type IS 'Type of identification document: national_id (Thai ID card) or passport (Foreign passport)';


--
-- Name: COLUMN face_verifications.document_id; Type: COMMENT; Schema: public; Owner: admin
--

COMMENT ON COLUMN public.face_verifications.document_id IS 'References provider_documents.document_id for the ID card or passport document';


--
-- Name: face_verifications_verification_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.face_verifications_verification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.face_verifications_verification_id_seq OWNER TO admin;

--
-- Name: face_verifications_verification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.face_verifications_verification_id_seq OWNED BY public.face_verifications.verification_id;


--
-- Name: favorites; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.favorites (
    favorite_id integer NOT NULL,
    client_id integer NOT NULL,
    provider_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.favorites OWNER TO admin;

--
-- Name: favorites_favorite_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.favorites_favorite_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.favorites_favorite_id_seq OWNER TO admin;

--
-- Name: favorites_favorite_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.favorites_favorite_id_seq OWNED BY public.favorites.favorite_id;


--
-- Name: genders; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.genders (
    gender_id integer NOT NULL,
    gender_name character varying(50) NOT NULL
);


ALTER TABLE public.genders OWNER TO admin;

--
-- Name: genders_gender_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.genders_gender_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.genders_gender_id_seq OWNER TO admin;

--
-- Name: genders_gender_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.genders_gender_id_seq OWNED BY public.genders.gender_id;


--
-- Name: god_commission_balance; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.god_commission_balance (
    balance_id integer NOT NULL,
    god_user_id integer NOT NULL,
    platform_bank_account_id integer NOT NULL,
    total_commission_collected numeric(12,2) DEFAULT 0.00 NOT NULL,
    total_transferred numeric(12,2) DEFAULT 0.00 NOT NULL,
    current_balance numeric(12,2) DEFAULT 0.00 NOT NULL,
    total_withdrawals_processed integer DEFAULT 0,
    average_withdrawal_amount numeric(12,2) DEFAULT 0.00,
    last_updated timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT positive_balances CHECK (((total_commission_collected >= (0)::numeric) AND (total_transferred >= (0)::numeric) AND (current_balance >= (0)::numeric)))
);


ALTER TABLE public.god_commission_balance OWNER TO admin;

--
-- Name: god_commission_balance_balance_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.god_commission_balance_balance_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.god_commission_balance_balance_id_seq OWNER TO admin;

--
-- Name: god_commission_balance_balance_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.god_commission_balance_balance_id_seq OWNED BY public.god_commission_balance.balance_id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.messages (
    message_id integer NOT NULL,
    conversation_id integer NOT NULL,
    sender_id integer NOT NULL,
    content text NOT NULL,
    is_read boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.messages OWNER TO admin;

--
-- Name: messages_message_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.messages_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.messages_message_id_seq OWNER TO admin;

--
-- Name: messages_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.messages_message_id_seq OWNED BY public.messages.message_id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.notifications (
    notification_id integer NOT NULL,
    user_id integer NOT NULL,
    type character varying(50) NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    is_read boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.notifications OWNER TO admin;

--
-- Name: notifications_notification_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.notifications_notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notifications_notification_id_seq OWNER TO admin;

--
-- Name: notifications_notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.notifications_notification_id_seq OWNED BY public.notifications.notification_id;


--
-- Name: platform_bank_accounts; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.platform_bank_accounts (
    platform_bank_id integer NOT NULL,
    bank_name character varying(100) NOT NULL,
    bank_code character varying(10),
    account_number character varying(50) NOT NULL,
    account_name character varying(200) NOT NULL,
    account_type character varying(20) DEFAULT 'current'::character varying,
    branch_name character varying(100),
    account_holder character varying(200),
    account_holder_id_card character varying(50),
    current_balance numeric(12,2) DEFAULT 0.00,
    total_inflow numeric(12,2) DEFAULT 0.00,
    total_outflow numeric(12,2) DEFAULT 0.00,
    is_active boolean DEFAULT true,
    is_default boolean DEFAULT false,
    owned_by integer,
    notes text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.platform_bank_accounts OWNER TO admin;

--
-- Name: platform_bank_accounts_platform_bank_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.platform_bank_accounts_platform_bank_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.platform_bank_accounts_platform_bank_id_seq OWNER TO admin;

--
-- Name: platform_bank_accounts_platform_bank_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.platform_bank_accounts_platform_bank_id_seq OWNED BY public.platform_bank_accounts.platform_bank_id;


--
-- Name: provider_availability; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_availability (
    availability_id integer NOT NULL,
    provider_id integer NOT NULL,
    day_of_week integer NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    is_active boolean DEFAULT true,
    CONSTRAINT provider_availability_day_of_week_check CHECK (((day_of_week >= 0) AND (day_of_week <= 6)))
);


ALTER TABLE public.provider_availability OWNER TO admin;

--
-- Name: provider_availability_availability_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.provider_availability_availability_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provider_availability_availability_id_seq OWNER TO admin;

--
-- Name: provider_availability_availability_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.provider_availability_availability_id_seq OWNED BY public.provider_availability.availability_id;


--
-- Name: provider_categories; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_categories (
    provider_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.provider_categories OWNER TO admin;

--
-- Name: provider_documents; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_documents (
    document_id integer NOT NULL,
    user_id integer NOT NULL,
    document_type character varying(50) NOT NULL,
    document_url text NOT NULL,
    verification_status character varying(50) DEFAULT 'pending'::character varying,
    uploaded_at timestamp with time zone DEFAULT now(),
    verified_at timestamp with time zone,
    admin_notes text,
    requires_face_match boolean DEFAULT false
);


ALTER TABLE public.provider_documents OWNER TO admin;

--
-- Name: provider_documents_document_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.provider_documents_document_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provider_documents_document_id_seq OWNER TO admin;

--
-- Name: provider_documents_document_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.provider_documents_document_id_seq OWNED BY public.provider_documents.document_id;


--
-- Name: provider_fee_notifications; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_fee_notifications (
    notification_id integer NOT NULL,
    provider_id integer NOT NULL,
    platform_rate numeric(5,4) NOT NULL,
    payment_gateway_rate numeric(5,4) NOT NULL,
    total_rate numeric(5,4) NOT NULL,
    notification_type character varying(50) NOT NULL,
    shown_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    acknowledged boolean DEFAULT false,
    acknowledged_at timestamp without time zone,
    notification_channel character varying(50),
    notes text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.provider_fee_notifications OWNER TO admin;

--
-- Name: provider_fee_notifications_notification_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.provider_fee_notifications_notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provider_fee_notifications_notification_id_seq OWNER TO admin;

--
-- Name: provider_fee_notifications_notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.provider_fee_notifications_notification_id_seq OWNED BY public.provider_fee_notifications.notification_id;


--
-- Name: provider_schedules; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_schedules (
    schedule_id integer NOT NULL,
    provider_id integer NOT NULL,
    booking_id integer,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone NOT NULL,
    status character varying(20) DEFAULT 'available'::character varying NOT NULL,
    location_type character varying(20),
    location_address text,
    location_province character varying(100),
    location_district character varying(100),
    latitude numeric(10,8),
    longitude numeric(11,8),
    notes text,
    is_visible_to_admin boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT provider_schedules_location_type_check CHECK (((location_type)::text = ANY ((ARRAY['Incall'::character varying, 'Outcall'::character varying, 'Both'::character varying])::text[]))),
    CONSTRAINT provider_schedules_status_check CHECK (((status)::text = ANY ((ARRAY['available'::character varying, 'booked'::character varying, 'blocked'::character varying])::text[])))
);


ALTER TABLE public.provider_schedules OWNER TO admin;

--
-- Name: provider_schedules_schedule_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.provider_schedules_schedule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provider_schedules_schedule_id_seq OWNER TO admin;

--
-- Name: provider_schedules_schedule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.provider_schedules_schedule_id_seq OWNED BY public.provider_schedules.schedule_id;


--
-- Name: provider_tier_history; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.provider_tier_history (
    history_id integer NOT NULL,
    provider_id integer NOT NULL,
    old_tier_id integer,
    new_tier_id integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now(),
    reason text
);


ALTER TABLE public.provider_tier_history OWNER TO admin;

--
-- Name: provider_tier_history_history_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.provider_tier_history_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provider_tier_history_history_id_seq OWNER TO admin;

--
-- Name: provider_tier_history_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.provider_tier_history_history_id_seq OWNED BY public.provider_tier_history.history_id;


--
-- Name: reports; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.reports (
    report_id integer NOT NULL,
    reporter_id integer NOT NULL,
    reported_user_id integer NOT NULL,
    reason character varying(255) NOT NULL,
    description text,
    status character varying(50) DEFAULT 'pending'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    resolved_at timestamp with time zone,
    admin_notes text
);


ALTER TABLE public.reports OWNER TO admin;

--
-- Name: reports_report_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.reports_report_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.reports_report_id_seq OWNER TO admin;

--
-- Name: reports_report_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.reports_report_id_seq OWNED BY public.reports.report_id;


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.reviews (
    review_id integer NOT NULL,
    booking_id integer NOT NULL,
    client_id integer NOT NULL,
    provider_id integer NOT NULL,
    rating integer NOT NULL,
    comment text,
    is_verified boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE public.reviews OWNER TO admin;

--
-- Name: reviews_review_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.reviews_review_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.reviews_review_id_seq OWNER TO admin;

--
-- Name: reviews_review_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.reviews_review_id_seq OWNED BY public.reviews.review_id;


--
-- Name: service_categories; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.service_categories (
    category_id integer NOT NULL,
    name character varying(100) NOT NULL,
    icon character varying(50),
    description text,
    created_at timestamp with time zone DEFAULT now(),
    name_thai character varying(100),
    is_adult boolean DEFAULT false,
    display_order integer DEFAULT 0,
    is_active boolean DEFAULT true
);


ALTER TABLE public.service_categories OWNER TO admin;

--
-- Name: service_categories_category_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.service_categories_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.service_categories_category_id_seq OWNER TO admin;

--
-- Name: service_categories_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.service_categories_category_id_seq OWNED BY public.service_categories.category_id;


--
-- Name: service_packages; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.service_packages (
    package_id integer NOT NULL,
    provider_id integer NOT NULL,
    package_name character varying(100) NOT NULL,
    description text,
    duration integer NOT NULL,
    price numeric(10,2) NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.service_packages OWNER TO admin;

--
-- Name: service_packages_package_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.service_packages_package_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.service_packages_package_id_seq OWNER TO admin;

--
-- Name: service_packages_package_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.service_packages_package_id_seq OWNED BY public.service_packages.package_id;


--
-- Name: tiers; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.tiers (
    tier_id integer NOT NULL,
    name character varying(50) NOT NULL,
    access_level integer NOT NULL,
    price_monthly numeric(10,2) DEFAULT 0.00 NOT NULL
);


ALTER TABLE public.tiers OWNER TO admin;

--
-- Name: tiers_tier_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.tiers_tier_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tiers_tier_id_seq OWNER TO admin;

--
-- Name: tiers_tier_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.tiers_tier_id_seq OWNED BY public.tiers.tier_id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.transactions (
    transaction_id integer NOT NULL,
    wallet_id integer NOT NULL,
    booking_id integer,
    type character varying(50) NOT NULL,
    amount numeric(10,2) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying,
    stripe_transaction_id character varying(255),
    platform_fee numeric(10,2) DEFAULT 0.00,
    stripe_fee numeric(10,2) DEFAULT 0.00,
    net_amount numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    platform_commission numeric(12,2) DEFAULT 0.00,
    total_fee_percentage numeric(5,4) DEFAULT 0.1275
);


ALTER TABLE public.transactions OWNER TO admin;

--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.transactions_transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.transactions_transaction_id_seq OWNER TO admin;

--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.transactions_transaction_id_seq OWNED BY public.transactions.transaction_id;


--
-- Name: user_photos; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.user_photos (
    photo_id integer NOT NULL,
    user_id integer NOT NULL,
    photo_url text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    uploaded_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.user_photos OWNER TO admin;

--
-- Name: user_photos_photo_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.user_photos_photo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_photos_photo_id_seq OWNER TO admin;

--
-- Name: user_photos_photo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.user_photos_photo_id_seq OWNED BY public.user_photos.photo_id;


--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.user_profiles (
    user_id integer NOT NULL,
    bio text,
    location character varying(255),
    skills text[],
    profile_image_url text,
    updated_at timestamp with time zone DEFAULT now(),
    age integer,
    height integer,
    weight integer,
    ethnicity character varying(50),
    languages text[],
    working_hours character varying(100),
    is_available boolean DEFAULT false,
    service_type character varying(20)
);


ALTER TABLE public.user_profiles OWNER TO admin;

--
-- Name: user_verifications; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.user_verifications (
    verification_id integer NOT NULL,
    user_id integer NOT NULL,
    national_id_url text,
    health_cert_url text,
    face_scan_url text,
    submitted_at timestamp with time zone
);


ALTER TABLE public.user_verifications OWNER TO admin;

--
-- Name: user_verifications_verification_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.user_verifications_verification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_verifications_verification_id_seq OWNER TO admin;

--
-- Name: user_verifications_verification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.user_verifications_verification_id_seq OWNED BY public.user_verifications.verification_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    username character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash text,
    gender_id integer DEFAULT 4 NOT NULL,
    first_name character varying(100),
    last_name character varying(100),
    registration_date timestamp with time zone DEFAULT now(),
    google_id text,
    google_profile_picture text,
    tier_id integer DEFAULT 1,
    phone_number character varying(20),
    verification_status character varying(20) DEFAULT 'unverified'::character varying NOT NULL,
    is_admin boolean DEFAULT false NOT NULL,
    provider_level_id integer DEFAULT 1,
    face_verified boolean DEFAULT false,
    face_verification_id integer,
    profile_picture_url text
);


ALTER TABLE public.users OWNER TO admin;

--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_user_id_seq OWNER TO admin;

--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- Name: wallets; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.wallets (
    wallet_id integer NOT NULL,
    user_id integer NOT NULL,
    available_balance numeric(10,2) DEFAULT 0.00,
    pending_balance numeric(10,2) DEFAULT 0.00,
    total_earned numeric(10,2) DEFAULT 0.00,
    total_withdrawn numeric(10,2) DEFAULT 0.00,
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.wallets OWNER TO admin;

--
-- Name: wallets_wallet_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.wallets_wallet_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.wallets_wallet_id_seq OWNER TO admin;

--
-- Name: wallets_wallet_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.wallets_wallet_id_seq OWNED BY public.wallets.wallet_id;


--
-- Name: withdrawal_requests; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.withdrawal_requests (
    withdrawal_id integer NOT NULL,
    wallet_id integer NOT NULL,
    bank_account_id integer NOT NULL,
    amount numeric(10,2) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying,
    requested_at timestamp with time zone DEFAULT now(),
    approved_at timestamp with time zone,
    rejected_at timestamp with time zone,
    completed_at timestamp with time zone,
    admin_notes text,
    transfer_slip_url text
);


ALTER TABLE public.withdrawal_requests OWNER TO admin;

--
-- Name: withdrawal_requests_withdrawal_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.withdrawal_requests_withdrawal_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.withdrawal_requests_withdrawal_id_seq OWNER TO admin;

--
-- Name: withdrawal_requests_withdrawal_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.withdrawal_requests_withdrawal_id_seq OWNED BY public.withdrawal_requests.withdrawal_id;


--
-- Name: bank_accounts bank_account_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bank_accounts ALTER COLUMN bank_account_id SET DEFAULT nextval('public.bank_accounts_bank_account_id_seq'::regclass);


--
-- Name: blocks block_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.blocks ALTER COLUMN block_id SET DEFAULT nextval('public.blocks_block_id_seq'::regclass);


--
-- Name: bookings booking_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bookings ALTER COLUMN booking_id SET DEFAULT nextval('public.bookings_booking_id_seq'::regclass);


--
-- Name: commission_transactions commission_txn_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions ALTER COLUMN commission_txn_id SET DEFAULT nextval('public.commission_transactions_commission_txn_id_seq'::regclass);


--
-- Name: conversations conversation_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.conversations ALTER COLUMN conversation_id SET DEFAULT nextval('public.conversations_conversation_id_seq'::regclass);


--
-- Name: face_verifications verification_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications ALTER COLUMN verification_id SET DEFAULT nextval('public.face_verifications_verification_id_seq'::regclass);


--
-- Name: favorites favorite_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.favorites ALTER COLUMN favorite_id SET DEFAULT nextval('public.favorites_favorite_id_seq'::regclass);


--
-- Name: genders gender_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.genders ALTER COLUMN gender_id SET DEFAULT nextval('public.genders_gender_id_seq'::regclass);


--
-- Name: god_commission_balance balance_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.god_commission_balance ALTER COLUMN balance_id SET DEFAULT nextval('public.god_commission_balance_balance_id_seq'::regclass);


--
-- Name: messages message_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.messages ALTER COLUMN message_id SET DEFAULT nextval('public.messages_message_id_seq'::regclass);


--
-- Name: notifications notification_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.notifications ALTER COLUMN notification_id SET DEFAULT nextval('public.notifications_notification_id_seq'::regclass);


--
-- Name: platform_bank_accounts platform_bank_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.platform_bank_accounts ALTER COLUMN platform_bank_id SET DEFAULT nextval('public.platform_bank_accounts_platform_bank_id_seq'::regclass);


--
-- Name: provider_availability availability_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_availability ALTER COLUMN availability_id SET DEFAULT nextval('public.provider_availability_availability_id_seq'::regclass);


--
-- Name: provider_documents document_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_documents ALTER COLUMN document_id SET DEFAULT nextval('public.provider_documents_document_id_seq'::regclass);


--
-- Name: provider_fee_notifications notification_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_fee_notifications ALTER COLUMN notification_id SET DEFAULT nextval('public.provider_fee_notifications_notification_id_seq'::regclass);


--
-- Name: provider_schedules schedule_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_schedules ALTER COLUMN schedule_id SET DEFAULT nextval('public.provider_schedules_schedule_id_seq'::regclass);


--
-- Name: provider_tier_history history_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_tier_history ALTER COLUMN history_id SET DEFAULT nextval('public.provider_tier_history_history_id_seq'::regclass);


--
-- Name: reports report_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reports ALTER COLUMN report_id SET DEFAULT nextval('public.reports_report_id_seq'::regclass);


--
-- Name: reviews review_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews ALTER COLUMN review_id SET DEFAULT nextval('public.reviews_review_id_seq'::regclass);


--
-- Name: service_categories category_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_categories ALTER COLUMN category_id SET DEFAULT nextval('public.service_categories_category_id_seq'::regclass);


--
-- Name: service_packages package_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_packages ALTER COLUMN package_id SET DEFAULT nextval('public.service_packages_package_id_seq'::regclass);


--
-- Name: tiers tier_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.tiers ALTER COLUMN tier_id SET DEFAULT nextval('public.tiers_tier_id_seq'::regclass);


--
-- Name: transactions transaction_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.transactions ALTER COLUMN transaction_id SET DEFAULT nextval('public.transactions_transaction_id_seq'::regclass);


--
-- Name: user_photos photo_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_photos ALTER COLUMN photo_id SET DEFAULT nextval('public.user_photos_photo_id_seq'::regclass);


--
-- Name: user_verifications verification_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_verifications ALTER COLUMN verification_id SET DEFAULT nextval('public.user_verifications_verification_id_seq'::regclass);


--
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Name: wallets wallet_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.wallets ALTER COLUMN wallet_id SET DEFAULT nextval('public.wallets_wallet_id_seq'::regclass);


--
-- Name: withdrawal_requests withdrawal_id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.withdrawal_requests ALTER COLUMN withdrawal_id SET DEFAULT nextval('public.withdrawal_requests_withdrawal_id_seq'::regclass);


--
-- Data for Name: bank_accounts; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.bank_accounts (bank_account_id, user_id, bank_name, account_number, account_holder_name, is_default, created_at) FROM stdin;
\.


--
-- Data for Name: blocks; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.blocks (block_id, blocker_id, blocked_id, created_at) FROM stdin;
\.


--
-- Data for Name: bookings; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.bookings (booking_id, client_id, provider_id, package_id, booking_date, start_time, end_time, total_price, status, location, special_notes, created_at, updated_at, completed_at, cancelled_at, cancellation_reason) FROM stdin;
\.


--
-- Data for Name: commission_transactions; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.commission_transactions (commission_txn_id, booking_id, transaction_id, booking_amount, commission_rate, commission_amount, provider_amount, provider_id, platform_bank_account_id, status, collected_at, refunded_at, refund_reason, notes, created_at) FROM stdin;
\.


--
-- Data for Name: conversations; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.conversations (conversation_id, user1_id, user2_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: face_verifications; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.face_verifications (verification_id, user_id, selfie_url, liveness_video_url, match_confidence, is_match, national_id_photo_url, liveness_passed, liveness_confidence, verification_status, api_provider, api_response_data, created_at, verified_at, verified_by, rejection_reason, retry_count, document_type, document_id) FROM stdin;
\.


--
-- Data for Name: favorites; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.favorites (favorite_id, client_id, provider_id, created_at) FROM stdin;
\.


--
-- Data for Name: genders; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.genders (gender_id, gender_name) FROM stdin;
1	Male
2	Female
3	Other
4	Prefer not to say
\.


--
-- Data for Name: god_commission_balance; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.god_commission_balance (balance_id, god_user_id, platform_bank_account_id, total_commission_collected, total_transferred, current_balance, total_withdrawals_processed, average_withdrawal_amount, last_updated, created_at) FROM stdin;
1	1	9	0.00	0.00	0.00	0	0.00	2025-12-02 07:13:09.203888	2025-12-02 07:13:09.203888
\.


--
-- Data for Name: messages; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.messages (message_id, conversation_id, sender_id, content, is_read, created_at) FROM stdin;
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.notifications (notification_id, user_id, type, title, message, is_read, created_at) FROM stdin;
\.


--
-- Data for Name: platform_bank_accounts; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.platform_bank_accounts (platform_bank_id, bank_name, bank_code, account_number, account_name, account_type, branch_name, account_holder, account_holder_id_card, current_balance, total_inflow, total_outflow, is_active, is_default, owned_by, notes, created_at, updated_at) FROM stdin;
9	ธนาคารกสิกรไทย	KBANK	XXX-X-XXXXX-X	บริษัท SkillMatch จำกัด	current	สาขาสีลม	นาย GOD Master	\N	0.00	0.00	0.00	t	t	1	บัญชีธนาคารหลักของแพลตฟอร์ม ใช้สำหรับโอนเงินให้ providers ทั้งหมด	2025-12-02 07:13:08.626702	2025-12-02 07:13:08.626702
\.


--
-- Data for Name: provider_availability; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_availability (availability_id, provider_id, day_of_week, start_time, end_time, is_active) FROM stdin;
\.


--
-- Data for Name: provider_categories; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_categories (provider_id, category_id, created_at) FROM stdin;
\.


--
-- Data for Name: provider_documents; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_documents (document_id, user_id, document_type, document_url, verification_status, uploaded_at, verified_at, admin_notes, requires_face_match) FROM stdin;
\.


--
-- Data for Name: provider_fee_notifications; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_fee_notifications (notification_id, provider_id, platform_rate, payment_gateway_rate, total_rate, notification_type, shown_at, acknowledged, acknowledged_at, notification_channel, notes, created_at) FROM stdin;
\.


--
-- Data for Name: provider_schedules; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_schedules (schedule_id, provider_id, booking_id, start_time, end_time, status, location_type, location_address, location_province, location_district, latitude, longitude, notes, is_visible_to_admin, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: provider_tier_history; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.provider_tier_history (history_id, provider_id, old_tier_id, new_tier_id, changed_at, reason) FROM stdin;
\.


--
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.reports (report_id, reporter_id, reported_user_id, reason, description, status, created_at, resolved_at, admin_notes) FROM stdin;
\.


--
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.reviews (review_id, booking_id, client_id, provider_id, rating, comment, is_verified, created_at) FROM stdin;
\.


--
-- Data for Name: service_categories; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.service_categories (category_id, name, icon, description, created_at, name_thai, is_adult, display_order, is_active) FROM stdin;
1	Massage	💆	Professional massage services	2025-12-02 09:43:03.608215+00	นวด	f	1	t
2	Spa	🧖	Spa and wellness treatments	2025-12-02 09:43:03.608215+00	สปา	f	2	t
3	Beauty	💄	Beauty and cosmetic services	2025-12-02 09:43:03.608215+00	ความงาม	f	3	t
4	Wellness	🧘	Health and wellness services	2025-12-02 09:43:03.608215+00	สุขภาพ	f	4	t
5	Therapy	🩺	Therapeutic services	2025-12-02 09:43:03.608215+00	บำบัด	f	5	t
\.


--
-- Data for Name: service_packages; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.service_packages (package_id, provider_id, package_name, description, duration, price, is_active, created_at) FROM stdin;
\.


--
-- Data for Name: tiers; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.tiers (tier_id, name, access_level, price_monthly) FROM stdin;
1	General	0	0.00
2	Silver	1	9.99
3	Diamond	2	29.99
4	Premium	3	99.99
5	GOD	999	9999.99
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.transactions (transaction_id, wallet_id, booking_id, type, amount, status, stripe_transaction_id, platform_fee, stripe_fee, net_amount, created_at, platform_commission, total_fee_percentage) FROM stdin;
\.


--
-- Data for Name: user_photos; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.user_photos (photo_id, user_id, photo_url, sort_order, uploaded_at) FROM stdin;
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.user_profiles (user_id, bio, location, skills, profile_image_url, updated_at, age, height, weight, ethnicity, languages, working_hours, is_available, service_type) FROM stdin;
1	\N	\N	\N	\N	2025-12-02 07:04:21.833114+00	\N	\N	\N	\N	\N	\N	f	\N
\.


--
-- Data for Name: user_verifications; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.user_verifications (verification_id, user_id, national_id_url, health_cert_url, face_scan_url, submitted_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.users (user_id, username, email, password_hash, gender_id, first_name, last_name, registration_date, google_id, google_profile_picture, tier_id, phone_number, verification_status, is_admin, provider_level_id, face_verified, face_verification_id, profile_picture_url) FROM stdin;
1	The BOB Film	audikoratair@gmail.com	\N	4	The BOB	Film	2025-12-02 07:04:21.764211+00	103537582873738046632	https://lh3.googleusercontent.com/a/ACg8ocKEs-hSuidV2Wk-9WI6PNnxCZAUV3RbBSZ7Ac9peloo8N7k3po=s96-c	5	\N	verified	t	1	f	\N	https://lh3.googleusercontent.com/a/ACg8ocKEs-hSuidV2Wk-9WI6PNnxCZAUV3RbBSZ7Ac9peloo8N7k3po=s96-c
\.


--
-- Data for Name: wallets; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.wallets (wallet_id, user_id, available_balance, pending_balance, total_earned, total_withdrawn, updated_at) FROM stdin;
\.


--
-- Data for Name: withdrawal_requests; Type: TABLE DATA; Schema: public; Owner: admin
--

COPY public.withdrawal_requests (withdrawal_id, wallet_id, bank_account_id, amount, status, requested_at, approved_at, rejected_at, completed_at, admin_notes, transfer_slip_url) FROM stdin;
\.


--
-- Name: bank_accounts_bank_account_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.bank_accounts_bank_account_id_seq', 1, false);


--
-- Name: blocks_block_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.blocks_block_id_seq', 1, false);


--
-- Name: bookings_booking_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.bookings_booking_id_seq', 1, false);


--
-- Name: commission_transactions_commission_txn_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.commission_transactions_commission_txn_id_seq', 1, false);


--
-- Name: conversations_conversation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.conversations_conversation_id_seq', 1, false);


--
-- Name: face_verifications_verification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.face_verifications_verification_id_seq', 1, false);


--
-- Name: favorites_favorite_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.favorites_favorite_id_seq', 1, false);


--
-- Name: genders_gender_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.genders_gender_id_seq', 1, false);


--
-- Name: god_commission_balance_balance_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.god_commission_balance_balance_id_seq', 16, true);


--
-- Name: messages_message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.messages_message_id_seq', 1, false);


--
-- Name: notifications_notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.notifications_notification_id_seq', 1, false);


--
-- Name: platform_bank_accounts_platform_bank_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.platform_bank_accounts_platform_bank_id_seq', 24, true);


--
-- Name: provider_availability_availability_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.provider_availability_availability_id_seq', 1, false);


--
-- Name: provider_documents_document_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.provider_documents_document_id_seq', 1, false);


--
-- Name: provider_fee_notifications_notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.provider_fee_notifications_notification_id_seq', 1, false);


--
-- Name: provider_schedules_schedule_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.provider_schedules_schedule_id_seq', 1, false);


--
-- Name: provider_tier_history_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.provider_tier_history_history_id_seq', 1, false);


--
-- Name: reports_report_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.reports_report_id_seq', 1, false);


--
-- Name: reviews_review_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.reviews_review_id_seq', 1, false);


--
-- Name: service_categories_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.service_categories_category_id_seq', 50, true);


--
-- Name: service_packages_package_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.service_packages_package_id_seq', 1, false);


--
-- Name: tiers_tier_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.tiers_tier_id_seq', 1, false);


--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.transactions_transaction_id_seq', 1, false);


--
-- Name: user_photos_photo_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.user_photos_photo_id_seq', 1, false);


--
-- Name: user_verifications_verification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.user_verifications_verification_id_seq', 1, false);


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.users_user_id_seq', 1, true);


--
-- Name: wallets_wallet_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.wallets_wallet_id_seq', 1, false);


--
-- Name: withdrawal_requests_withdrawal_id_seq; Type: SEQUENCE SET; Schema: public; Owner: admin
--

SELECT pg_catalog.setval('public.withdrawal_requests_withdrawal_id_seq', 1, false);


--
-- Name: bank_accounts bank_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bank_accounts
    ADD CONSTRAINT bank_accounts_pkey PRIMARY KEY (bank_account_id);


--
-- Name: blocks blocks_blocker_id_blocked_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_blocker_id_blocked_id_key UNIQUE (blocker_id, blocked_id);


--
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (block_id);


--
-- Name: bookings bookings_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_pkey PRIMARY KEY (booking_id);


--
-- Name: commission_transactions commission_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions
    ADD CONSTRAINT commission_transactions_pkey PRIMARY KEY (commission_txn_id);


--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (conversation_id);


--
-- Name: conversations conversations_user1_id_user2_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_user1_id_user2_id_key UNIQUE (user1_id, user2_id);


--
-- Name: face_verifications face_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT face_verifications_pkey PRIMARY KEY (verification_id);


--
-- Name: favorites favorites_client_id_provider_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_client_id_provider_id_key UNIQUE (client_id, provider_id);


--
-- Name: favorites favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_pkey PRIMARY KEY (favorite_id);


--
-- Name: genders genders_gender_name_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.genders
    ADD CONSTRAINT genders_gender_name_key UNIQUE (gender_name);


--
-- Name: genders genders_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.genders
    ADD CONSTRAINT genders_pkey PRIMARY KEY (gender_id);


--
-- Name: god_commission_balance god_commission_balance_god_user_id_platform_bank_account_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.god_commission_balance
    ADD CONSTRAINT god_commission_balance_god_user_id_platform_bank_account_id_key UNIQUE (god_user_id, platform_bank_account_id);


--
-- Name: god_commission_balance god_commission_balance_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.god_commission_balance
    ADD CONSTRAINT god_commission_balance_pkey PRIMARY KEY (balance_id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);


--
-- Name: platform_bank_accounts platform_bank_accounts_account_number_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.platform_bank_accounts
    ADD CONSTRAINT platform_bank_accounts_account_number_key UNIQUE (account_number);


--
-- Name: platform_bank_accounts platform_bank_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.platform_bank_accounts
    ADD CONSTRAINT platform_bank_accounts_pkey PRIMARY KEY (platform_bank_id);


--
-- Name: provider_availability provider_availability_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_availability
    ADD CONSTRAINT provider_availability_pkey PRIMARY KEY (availability_id);


--
-- Name: provider_availability provider_availability_provider_id_day_of_week_start_time_en_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_availability
    ADD CONSTRAINT provider_availability_provider_id_day_of_week_start_time_en_key UNIQUE (provider_id, day_of_week, start_time, end_time);


--
-- Name: provider_categories provider_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_categories
    ADD CONSTRAINT provider_categories_pkey PRIMARY KEY (provider_id, category_id);


--
-- Name: provider_documents provider_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_documents
    ADD CONSTRAINT provider_documents_pkey PRIMARY KEY (document_id);


--
-- Name: provider_fee_notifications provider_fee_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_fee_notifications
    ADD CONSTRAINT provider_fee_notifications_pkey PRIMARY KEY (notification_id);


--
-- Name: provider_schedules provider_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_schedules
    ADD CONSTRAINT provider_schedules_pkey PRIMARY KEY (schedule_id);


--
-- Name: provider_tier_history provider_tier_history_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_tier_history
    ADD CONSTRAINT provider_tier_history_pkey PRIMARY KEY (history_id);


--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (report_id);


--
-- Name: reviews reviews_booking_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_booking_id_key UNIQUE (booking_id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (review_id);


--
-- Name: service_categories service_categories_name_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_categories
    ADD CONSTRAINT service_categories_name_key UNIQUE (name);


--
-- Name: service_categories service_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (category_id);


--
-- Name: service_packages service_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_packages
    ADD CONSTRAINT service_packages_pkey PRIMARY KEY (package_id);


--
-- Name: tiers tiers_access_level_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.tiers
    ADD CONSTRAINT tiers_access_level_key UNIQUE (access_level);


--
-- Name: tiers tiers_name_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.tiers
    ADD CONSTRAINT tiers_name_key UNIQUE (name);


--
-- Name: tiers tiers_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.tiers
    ADD CONSTRAINT tiers_pkey PRIMARY KEY (tier_id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (transaction_id);


--
-- Name: user_photos user_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_photos
    ADD CONSTRAINT user_photos_pkey PRIMARY KEY (photo_id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id);


--
-- Name: user_verifications user_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_verifications
    ADD CONSTRAINT user_verifications_pkey PRIMARY KEY (verification_id);


--
-- Name: user_verifications user_verifications_user_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_verifications
    ADD CONSTRAINT user_verifications_user_id_key UNIQUE (user_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_google_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_google_id_key UNIQUE (google_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (wallet_id);


--
-- Name: wallets wallets_user_id_key; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_user_id_key UNIQUE (user_id);


--
-- Name: withdrawal_requests withdrawal_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.withdrawal_requests
    ADD CONSTRAINT withdrawal_requests_pkey PRIMARY KEY (withdrawal_id);


--
-- Name: email_idx; Type: INDEX; Schema: public; Owner: admin
--

CREATE UNIQUE INDEX email_idx ON public.users USING btree (email);


--
-- Name: google_id_idx; Type: INDEX; Schema: public; Owner: admin
--

CREATE UNIQUE INDEX google_id_idx ON public.users USING btree (google_id);


--
-- Name: idx_blocks_blocked; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_blocks_blocked ON public.blocks USING btree (blocked_id);


--
-- Name: idx_blocks_blocker; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_blocks_blocker ON public.blocks USING btree (blocker_id);


--
-- Name: idx_bookings_client; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_bookings_client ON public.bookings USING btree (client_id);


--
-- Name: idx_bookings_date; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_bookings_date ON public.bookings USING btree (booking_date);


--
-- Name: idx_bookings_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_bookings_provider ON public.bookings USING btree (provider_id);


--
-- Name: idx_bookings_status; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_bookings_status ON public.bookings USING btree (status);


--
-- Name: idx_commission_transactions_booking; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_commission_transactions_booking ON public.commission_transactions USING btree (booking_id);


--
-- Name: idx_commission_transactions_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_commission_transactions_provider ON public.commission_transactions USING btree (provider_id);


--
-- Name: idx_face_verifications_created_at; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_face_verifications_created_at ON public.face_verifications USING btree (created_at DESC);


--
-- Name: idx_face_verifications_document_type; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_face_verifications_document_type ON public.face_verifications USING btree (document_type);


--
-- Name: idx_face_verifications_status; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_face_verifications_status ON public.face_verifications USING btree (verification_status);


--
-- Name: idx_face_verifications_user_id; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_face_verifications_user_id ON public.face_verifications USING btree (user_id);


--
-- Name: idx_favorites_client; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_favorites_client ON public.favorites USING btree (client_id);


--
-- Name: idx_favorites_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_favorites_provider ON public.favorites USING btree (provider_id);


--
-- Name: idx_god_commission_balance_user; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_god_commission_balance_user ON public.god_commission_balance USING btree (god_user_id);


--
-- Name: idx_messages_conversation; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_messages_conversation ON public.messages USING btree (conversation_id);


--
-- Name: idx_messages_sender; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_messages_sender ON public.messages USING btree (sender_id);


--
-- Name: idx_notifications_unread; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_notifications_unread ON public.notifications USING btree (user_id, is_read);


--
-- Name: idx_notifications_user; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_notifications_user ON public.notifications USING btree (user_id);


--
-- Name: idx_provider_docs_user; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_provider_docs_user ON public.provider_documents USING btree (user_id);


--
-- Name: idx_provider_fee_notifications_acknowledged; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_provider_fee_notifications_acknowledged ON public.provider_fee_notifications USING btree (acknowledged) WHERE (acknowledged = false);


--
-- Name: idx_provider_fee_notifications_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_provider_fee_notifications_provider ON public.provider_fee_notifications USING btree (provider_id);


--
-- Name: idx_reports_status; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_reports_status ON public.reports USING btree (status);


--
-- Name: idx_reviews_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_reviews_provider ON public.reviews USING btree (provider_id);


--
-- Name: idx_schedules_booking; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_schedules_booking ON public.provider_schedules USING btree (booking_id);


--
-- Name: idx_schedules_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_schedules_provider ON public.provider_schedules USING btree (provider_id);


--
-- Name: idx_schedules_status; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_schedules_status ON public.provider_schedules USING btree (status);


--
-- Name: idx_schedules_time; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_schedules_time ON public.provider_schedules USING btree (start_time, end_time);


--
-- Name: idx_tier_history_provider; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_tier_history_provider ON public.provider_tier_history USING btree (provider_id);


--
-- Name: idx_transactions_wallet; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_transactions_wallet ON public.transactions USING btree (wallet_id);


--
-- Name: idx_withdrawal_status; Type: INDEX; Schema: public; Owner: admin
--

CREATE INDEX idx_withdrawal_status ON public.withdrawal_requests USING btree (status);


--
-- Name: provider_schedules trigger_update_schedule_timestamp; Type: TRIGGER; Schema: public; Owner: admin
--

CREATE TRIGGER trigger_update_schedule_timestamp BEFORE UPDATE ON public.provider_schedules FOR EACH ROW EXECUTE FUNCTION public.update_schedule_timestamp();


--
-- Name: face_verifications trigger_update_user_face_verification; Type: TRIGGER; Schema: public; Owner: admin
--

CREATE TRIGGER trigger_update_user_face_verification AFTER UPDATE ON public.face_verifications FOR EACH ROW EXECUTE FUNCTION public.update_user_face_verification();


--
-- Name: bank_accounts bank_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bank_accounts
    ADD CONSTRAINT bank_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: blocks blocks_blocked_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_blocked_id_fkey FOREIGN KEY (blocked_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: blocks blocks_blocker_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_blocker_id_fkey FOREIGN KEY (blocker_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: bookings bookings_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: bookings bookings_package_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_package_id_fkey FOREIGN KEY (package_id) REFERENCES public.service_packages(package_id);


--
-- Name: bookings bookings_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: commission_transactions commission_transactions_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions
    ADD CONSTRAINT commission_transactions_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(booking_id);


--
-- Name: commission_transactions commission_transactions_platform_bank_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions
    ADD CONSTRAINT commission_transactions_platform_bank_account_id_fkey FOREIGN KEY (platform_bank_account_id) REFERENCES public.platform_bank_accounts(platform_bank_id);


--
-- Name: commission_transactions commission_transactions_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions
    ADD CONSTRAINT commission_transactions_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id);


--
-- Name: commission_transactions commission_transactions_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.commission_transactions
    ADD CONSTRAINT commission_transactions_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(transaction_id);


--
-- Name: conversations conversations_user1_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_user1_id_fkey FOREIGN KEY (user1_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: conversations conversations_user2_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_user2_id_fkey FOREIGN KEY (user2_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: face_verifications face_verifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT face_verifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: face_verifications face_verifications_verified_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT face_verifications_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES public.users(user_id);


--
-- Name: favorites favorites_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: favorites favorites_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: face_verifications fk_document; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT fk_document FOREIGN KEY (document_id) REFERENCES public.provider_documents(document_id) ON DELETE SET NULL;


--
-- Name: face_verifications fk_user; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: face_verifications fk_verified_by; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.face_verifications
    ADD CONSTRAINT fk_verified_by FOREIGN KEY (verified_by) REFERENCES public.users(user_id);


--
-- Name: god_commission_balance god_commission_balance_god_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.god_commission_balance
    ADD CONSTRAINT god_commission_balance_god_user_id_fkey FOREIGN KEY (god_user_id) REFERENCES public.users(user_id);


--
-- Name: god_commission_balance god_commission_balance_platform_bank_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.god_commission_balance
    ADD CONSTRAINT god_commission_balance_platform_bank_account_id_fkey FOREIGN KEY (platform_bank_account_id) REFERENCES public.platform_bank_accounts(platform_bank_id);


--
-- Name: messages messages_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id) ON DELETE CASCADE;


--
-- Name: messages messages_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: platform_bank_accounts platform_bank_accounts_owned_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.platform_bank_accounts
    ADD CONSTRAINT platform_bank_accounts_owned_by_fkey FOREIGN KEY (owned_by) REFERENCES public.users(user_id);


--
-- Name: provider_availability provider_availability_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_availability
    ADD CONSTRAINT provider_availability_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: provider_categories provider_categories_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_categories
    ADD CONSTRAINT provider_categories_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.service_categories(category_id) ON DELETE CASCADE;


--
-- Name: provider_categories provider_categories_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_categories
    ADD CONSTRAINT provider_categories_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: provider_documents provider_documents_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_documents
    ADD CONSTRAINT provider_documents_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: provider_fee_notifications provider_fee_notifications_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_fee_notifications
    ADD CONSTRAINT provider_fee_notifications_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: provider_schedules provider_schedules_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_schedules
    ADD CONSTRAINT provider_schedules_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(booking_id) ON DELETE SET NULL;


--
-- Name: provider_schedules provider_schedules_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_schedules
    ADD CONSTRAINT provider_schedules_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: provider_tier_history provider_tier_history_new_tier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_tier_history
    ADD CONSTRAINT provider_tier_history_new_tier_id_fkey FOREIGN KEY (new_tier_id) REFERENCES public.tiers(tier_id);


--
-- Name: provider_tier_history provider_tier_history_old_tier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_tier_history
    ADD CONSTRAINT provider_tier_history_old_tier_id_fkey FOREIGN KEY (old_tier_id) REFERENCES public.tiers(tier_id);


--
-- Name: provider_tier_history provider_tier_history_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.provider_tier_history
    ADD CONSTRAINT provider_tier_history_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: reports reports_reported_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_reported_user_id_fkey FOREIGN KEY (reported_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: reports reports_reporter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: reviews reviews_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(booking_id) ON DELETE CASCADE;


--
-- Name: reviews reviews_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: reviews reviews_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: service_packages service_packages_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.service_packages
    ADD CONSTRAINT service_packages_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: transactions transactions_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.bookings(booking_id);


--
-- Name: transactions transactions_wallet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(wallet_id) ON DELETE CASCADE;


--
-- Name: user_photos user_photos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_photos
    ADD CONSTRAINT user_photos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: user_profiles user_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: user_verifications user_verifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.user_verifications
    ADD CONSTRAINT user_verifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: users users_face_verification_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_face_verification_id_fkey FOREIGN KEY (face_verification_id) REFERENCES public.face_verifications(verification_id);


--
-- Name: users users_gender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_gender_id_fkey FOREIGN KEY (gender_id) REFERENCES public.genders(gender_id);


--
-- Name: users users_provider_level_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_provider_level_id_fkey FOREIGN KEY (provider_level_id) REFERENCES public.tiers(tier_id);


--
-- Name: users users_tier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_tier_id_fkey FOREIGN KEY (tier_id) REFERENCES public.tiers(tier_id);


--
-- Name: wallets wallets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: withdrawal_requests withdrawal_requests_bank_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.withdrawal_requests
    ADD CONSTRAINT withdrawal_requests_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES public.bank_accounts(bank_account_id);


--
-- Name: withdrawal_requests withdrawal_requests_wallet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.withdrawal_requests
    ADD CONSTRAINT withdrawal_requests_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(wallet_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict QAjHrWmPLXykyZ4sBU13UNRfv1TH2cCBma9EeE8peTrynIw2Tm5pECq6iSD8Q18

