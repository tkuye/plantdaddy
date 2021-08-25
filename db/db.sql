--
-- PostgreSQL database dump
--

-- Dumped from database version 13.1
-- Dumped by pg_dump version 13.3

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth; Type: TABLE; Schema: public; Owner: plantdaddy
--

CREATE TABLE public.auth (
    id integer NOT NULL,
    username character varying(50) NOT NULL,
    password text NOT NULL,
    email character varying(100)
);


ALTER TABLE public.auth OWNER TO plantdaddy;

--
-- Name: auth_id_seq; Type: SEQUENCE; Schema: public; Owner: plantdaddy
--

CREATE SEQUENCE public.auth_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.auth_id_seq OWNER TO plantdaddy;

--
-- Name: auth_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: plantdaddy
--

ALTER SEQUENCE public.auth_id_seq OWNED BY public.auth.id;


--
-- Name: plant_data; Type: TABLE; Schema: public; Owner: plantdaddy
--

CREATE TABLE public.plant_data (
    id integer NOT NULL,
    device_id text,
    temperature double precision,
    humidity double precision,
    soil_moisture double precision,
    light double precision,
    "time" timestamp without time zone
);


ALTER TABLE public.plant_data OWNER TO plantdaddy;

--
-- Name: plant_data_id_seq; Type: SEQUENCE; Schema: public; Owner: plantdaddy
--

CREATE SEQUENCE public.plant_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.plant_data_id_seq OWNER TO plantdaddy;

--
-- Name: plant_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: plantdaddy
--

ALTER SEQUENCE public.plant_data_id_seq OWNED BY public.plant_data.id;


--
-- Name: registered_devices; Type: TABLE; Schema: public; Owner: plantdaddy
--

CREATE TABLE public.registered_devices (
    id integer NOT NULL,
    device_id text NOT NULL,
    register_date date NOT NULL,
    user_id integer,
    device_name text
);


ALTER TABLE public.registered_devices OWNER TO plantdaddy;

--
-- Name: registered_devices_id_seq; Type: SEQUENCE; Schema: public; Owner: plantdaddy
--

CREATE SEQUENCE public.registered_devices_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.registered_devices_id_seq OWNER TO plantdaddy;

--
-- Name: registered_devices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: plantdaddy
--

ALTER SEQUENCE public.registered_devices_id_seq OWNED BY public.registered_devices.id;


--
-- Name: session; Type: TABLE; Schema: public; Owner: plantdaddy
--

CREATE TABLE public.session (
    session_id text,
    usage integer,
    usage_time timestamp without time zone,
    id integer NOT NULL,
    device_id text
);


ALTER TABLE public.session OWNER TO plantdaddy;

--
-- Name: session_id_seq; Type: SEQUENCE; Schema: public; Owner: plantdaddy
--

CREATE SEQUENCE public.session_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.session_id_seq OWNER TO plantdaddy;

--
-- Name: session_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: plantdaddy
--

ALTER SEQUENCE public.session_id_seq OWNED BY public.session.id;


--
-- Name: auth id; Type: DEFAULT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.auth ALTER COLUMN id SET DEFAULT nextval('public.auth_id_seq'::regclass);


--
-- Name: plant_data id; Type: DEFAULT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.plant_data ALTER COLUMN id SET DEFAULT nextval('public.plant_data_id_seq'::regclass);


--
-- Name: registered_devices id; Type: DEFAULT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.registered_devices ALTER COLUMN id SET DEFAULT nextval('public.registered_devices_id_seq'::regclass);


--
-- Name: session id; Type: DEFAULT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.session ALTER COLUMN id SET DEFAULT nextval('public.session_id_seq'::regclass);


--
-- Name: auth auth_pkey; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.auth
    ADD CONSTRAINT auth_pkey PRIMARY KEY (id);


--
-- Name: plant_data plant_data_pkey; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.plant_data
    ADD CONSTRAINT plant_data_pkey PRIMARY KEY (id);


--
-- Name: registered_devices registered_devices_pkey; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.registered_devices
    ADD CONSTRAINT registered_devices_pkey PRIMARY KEY (device_id);


--
-- Name: session session_pkey; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.session
    ADD CONSTRAINT session_pkey PRIMARY KEY (id);


--
-- Name: session unique_device_id; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.session
    ADD CONSTRAINT unique_device_id UNIQUE (device_id);


--
-- Name: auth user_unique; Type: CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.auth
    ADD CONSTRAINT user_unique UNIQUE (username);


--
-- Name: session fk_device; Type: FK CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.session
    ADD CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES public.registered_devices(device_id) ON DELETE CASCADE;


--
-- Name: plant_data fk_device; Type: FK CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.plant_data
    ADD CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES public.registered_devices(device_id) ON DELETE CASCADE;


--
-- Name: registered_devices fk_user; Type: FK CONSTRAINT; Schema: public; Owner: plantdaddy
--

ALTER TABLE ONLY public.registered_devices
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.auth(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

