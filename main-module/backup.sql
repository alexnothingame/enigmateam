--
-- PostgreSQL database dump
--

\restrict IO3GToGZzgDZUllPraUypW0soItBsgfJo90BqfQOknhkaeLYuYJw5GilQ6C3EcR

-- Dumped from database version 16.11 (Debian 16.11-1.pgdg13+1)
-- Dumped by pg_dump version 16.11 (Debian 16.11-1.pgdg13+1)

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
-- Name: alembic_version; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.alembic_version (
    version_num character varying(32) NOT NULL
);


ALTER TABLE public.alembic_version OWNER TO app;

--
-- Name: answers; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.answers (
    id bigint NOT NULL,
    attempt_id bigint NOT NULL,
    question_id bigint NOT NULL,
    question_version integer NOT NULL,
    answer_index integer DEFAULT '-1'::integer NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_answers_idx CHECK ((answer_index >= '-1'::integer))
);


ALTER TABLE public.answers OWNER TO app;

--
-- Name: answers_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.answers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.answers_id_seq OWNER TO app;

--
-- Name: answers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.answers_id_seq OWNED BY public.answers.id;


--
-- Name: attempt_questions; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.attempt_questions (
    attempt_id bigint NOT NULL,
    question_id bigint NOT NULL,
    question_version integer NOT NULL,
    "position" integer NOT NULL,
    CONSTRAINT chk_aq_pos CHECK (("position" >= 0))
);


ALTER TABLE public.attempt_questions OWNER TO app;

--
-- Name: attempts; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.attempts (
    id bigint NOT NULL,
    test_id bigint NOT NULL,
    user_id bigint NOT NULL,
    status character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    CONSTRAINT chk_attempts_status CHECK (((status)::text = ANY ((ARRAY['in_progress'::character varying, 'finished'::character varying])::text[])))
);


ALTER TABLE public.attempts OWNER TO app;

--
-- Name: attempts_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.attempts_id_seq OWNER TO app;

--
-- Name: attempts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.attempts_id_seq OWNED BY public.attempts.id;


--
-- Name: course_enrollments; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.course_enrollments (
    course_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.course_enrollments OWNER TO app;

--
-- Name: courses; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.courses (
    id bigint NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    teacher_id bigint NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.courses OWNER TO app;

--
-- Name: courses_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.courses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.courses_id_seq OWNER TO app;

--
-- Name: courses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.courses_id_seq OWNED BY public.courses.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    payload text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notifications OWNER TO app;

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notifications_id_seq OWNER TO app;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: question_versions; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.question_versions (
    question_id bigint NOT NULL,
    version integer NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    options text[] NOT NULL,
    correct_index integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_qv_correct_ge0 CHECK ((correct_index >= 0)),
    CONSTRAINT chk_qv_correct_lt_len CHECK ((correct_index < array_length(options, 1))),
    CONSTRAINT chk_qv_opts_len CHECK ((array_length(options, 1) >= 2)),
    CONSTRAINT chk_qv_version CHECK ((version >= 1))
);


ALTER TABLE public.question_versions OWNER TO app;

--
-- Name: questions; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.questions (
    id bigint NOT NULL,
    author_id bigint NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.questions OWNER TO app;

--
-- Name: questions_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.questions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.questions_id_seq OWNER TO app;

--
-- Name: questions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.questions_id_seq OWNED BY public.questions.id;


--
-- Name: test_questions; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.test_questions (
    test_id bigint NOT NULL,
    question_id bigint NOT NULL,
    "position" integer NOT NULL,
    CONSTRAINT chk_tq_pos CHECK (("position" >= 0))
);


ALTER TABLE public.test_questions OWNER TO app;

--
-- Name: tests; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.tests (
    id bigint NOT NULL,
    course_id bigint NOT NULL,
    name text NOT NULL,
    author_id bigint NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tests OWNER TO app;

--
-- Name: tests_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.tests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tests_id_seq OWNER TO app;

--
-- Name: tests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.tests_id_seq OWNED BY public.tests.id;


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.user_roles (
    user_id bigint NOT NULL,
    role character varying NOT NULL,
    CONSTRAINT chk_user_roles_role CHECK (((role)::text = ANY ((ARRAY['student'::character varying, 'teacher'::character varying, 'admin'::character varying])::text[])))
);


ALTER TABLE public.user_roles OWNER TO app;

--
-- Name: users; Type: TABLE; Schema: public; Owner: app
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    full_name text NOT NULL,
    email text,
    is_blocked boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO app;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: app
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO app;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: app
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: answers id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.answers ALTER COLUMN id SET DEFAULT nextval('public.answers_id_seq'::regclass);


--
-- Name: attempts id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempts ALTER COLUMN id SET DEFAULT nextval('public.attempts_id_seq'::regclass);


--
-- Name: courses id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.courses ALTER COLUMN id SET DEFAULT nextval('public.courses_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: questions id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.questions ALTER COLUMN id SET DEFAULT nextval('public.questions_id_seq'::regclass);


--
-- Name: tests id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.tests ALTER COLUMN id SET DEFAULT nextval('public.tests_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: alembic_version; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.alembic_version (version_num) FROM stdin;
c4f5935ed29d
\.


--
-- Data for Name: answers; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.answers (id, attempt_id, question_id, question_version, answer_index, updated_at) FROM stdin;
\.


--
-- Data for Name: attempt_questions; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.attempt_questions (attempt_id, question_id, question_version, "position") FROM stdin;
\.


--
-- Data for Name: attempts; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.attempts (id, test_id, user_id, status, created_at, finished_at) FROM stdin;
1	500	20	in_progress	2026-01-09 13:52:12.125139+00	\N
\.


--
-- Data for Name: course_enrollments; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.course_enrollments (course_id, user_id, created_at) FROM stdin;
100	20	2026-01-09 13:45:43.613342+00
\.


--
-- Data for Name: courses; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.courses (id, name, description, teacher_id, is_deleted, created_at) FROM stdin;
100	Math	Test course	10	f	2026-01-09 13:27:23.272499+00
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.notifications (id, user_id, payload, created_at) FROM stdin;
\.


--
-- Data for Name: question_versions; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.question_versions (question_id, version, title, body, options, correct_index, created_at) FROM stdin;
700	1	Q1	Text	{A,B,C}	1	2026-01-09 13:52:04.300444+00
\.


--
-- Data for Name: questions; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.questions (id, author_id, is_deleted, created_at) FROM stdin;
700	30	f	2026-01-09 13:52:00.760536+00
\.


--
-- Data for Name: test_questions; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.test_questions (test_id, question_id, "position") FROM stdin;
\.


--
-- Data for Name: tests; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.tests (id, course_id, name, author_id, is_active, is_deleted, created_at) FROM stdin;
500	100	Quiz1	10	t	f	2026-01-09 13:51:52.807862+00
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.user_roles (user_id, role) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: app
--

COPY public.users (id, full_name, email, is_blocked, created_at) FROM stdin;
1	Test User	\N	f	2026-01-09 13:12:35.737024+00
10	Teacher T	\N	f	2026-01-09 13:26:55.094166+00
20	Student S	\N	f	2026-01-09 13:27:00.134936+00
30	Admin A	\N	f	2026-01-09 13:27:04.17705+00
\.


--
-- Name: answers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.answers_id_seq', 1, false);


--
-- Name: attempts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.attempts_id_seq', 1, true);


--
-- Name: courses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.courses_id_seq', 1, false);


--
-- Name: notifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.notifications_id_seq', 1, false);


--
-- Name: questions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.questions_id_seq', 1, false);


--
-- Name: tests_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.tests_id_seq', 1, false);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: app
--

SELECT pg_catalog.setval('public.users_id_seq', 1, false);


--
-- Name: alembic_version alembic_version_pkc; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.alembic_version
    ADD CONSTRAINT alembic_version_pkc PRIMARY KEY (version_num);


--
-- Name: answers answers_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.answers
    ADD CONSTRAINT answers_pkey PRIMARY KEY (id);


--
-- Name: attempt_questions attempt_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempt_questions
    ADD CONSTRAINT attempt_questions_pkey PRIMARY KEY (attempt_id, question_id);


--
-- Name: attempts attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempts
    ADD CONSTRAINT attempts_pkey PRIMARY KEY (id);


--
-- Name: course_enrollments course_enrollments_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.course_enrollments
    ADD CONSTRAINT course_enrollments_pkey PRIMARY KEY (course_id, user_id);


--
-- Name: courses courses_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: question_versions question_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.question_versions
    ADD CONSTRAINT question_versions_pkey PRIMARY KEY (question_id, version);


--
-- Name: questions questions_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_pkey PRIMARY KEY (id);


--
-- Name: test_questions test_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.test_questions
    ADD CONSTRAINT test_questions_pkey PRIMARY KEY (test_id, question_id);


--
-- Name: tests tests_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_pkey PRIMARY KEY (id);


--
-- Name: answers uq_answers_attempt_question; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.answers
    ADD CONSTRAINT uq_answers_attempt_question UNIQUE (attempt_id, question_id);


--
-- Name: attempt_questions uq_attempt_questions_position; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempt_questions
    ADD CONSTRAINT uq_attempt_questions_position UNIQUE (attempt_id, "position");


--
-- Name: attempts uq_attempts_test_user; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempts
    ADD CONSTRAINT uq_attempts_test_user UNIQUE (test_id, user_id);


--
-- Name: test_questions uq_test_questions_position; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.test_questions
    ADD CONSTRAINT uq_test_questions_position UNIQUE (test_id, "position");


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: answers answers_attempt_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.answers
    ADD CONSTRAINT answers_attempt_id_fkey FOREIGN KEY (attempt_id) REFERENCES public.attempts(id) ON DELETE CASCADE;


--
-- Name: attempt_questions attempt_questions_attempt_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempt_questions
    ADD CONSTRAINT attempt_questions_attempt_id_fkey FOREIGN KEY (attempt_id) REFERENCES public.attempts(id) ON DELETE CASCADE;


--
-- Name: attempts attempts_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempts
    ADD CONSTRAINT attempts_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.tests(id);


--
-- Name: attempts attempts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempts
    ADD CONSTRAINT attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: course_enrollments course_enrollments_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.course_enrollments
    ADD CONSTRAINT course_enrollments_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;


--
-- Name: course_enrollments course_enrollments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.course_enrollments
    ADD CONSTRAINT course_enrollments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: courses courses_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public.users(id);


--
-- Name: answers fk_answers_question_version; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.answers
    ADD CONSTRAINT fk_answers_question_version FOREIGN KEY (question_id, question_version) REFERENCES public.question_versions(question_id, version);


--
-- Name: attempt_questions fk_aq_question_version; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.attempt_questions
    ADD CONSTRAINT fk_aq_question_version FOREIGN KEY (question_id, question_version) REFERENCES public.question_versions(question_id, version);


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: question_versions question_versions_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.question_versions
    ADD CONSTRAINT question_versions_question_id_fkey FOREIGN KEY (question_id) REFERENCES public.questions(id) ON DELETE CASCADE;


--
-- Name: questions questions_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: test_questions test_questions_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.test_questions
    ADD CONSTRAINT test_questions_question_id_fkey FOREIGN KEY (question_id) REFERENCES public.questions(id);


--
-- Name: test_questions test_questions_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.test_questions
    ADD CONSTRAINT test_questions_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.tests(id) ON DELETE CASCADE;


--
-- Name: tests tests_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: tests tests_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id);


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: app
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict IO3GToGZzgDZUllPraUypW0soItBsgfJo90BqfQOknhkaeLYuYJw5GilQ6C3EcR

