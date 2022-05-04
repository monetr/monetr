package cache

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/1Password/srp"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SRPCache interface {
	// CacheAuthenticationSession takes the SRP object for the current authentication session and stores it somewhere.
	// If it is able to store it successfully then a sessionId will be returned that can be used to retrieve the
	// session from the cache.
	CacheAuthenticationSession(ctx context.Context, session *srp.SRP) (sessionId string, _ error)
	// LookupAuthenticationSession attempts to retrieve an SRP object from the cache for the provided sessionId, if the
	// provided sessionId is not valid; or if the session has expired, then an error will be returned.
	LookupAuthenticationSession(ctx context.Context, sessionId string) (*srp.SRP, error)
}

type srpCache struct {
	log   *logrus.Entry
	cache Cache
}

func NewSRPCache(log *logrus.Entry, cache Cache) SRPCache {
	return &srpCache{
		log:   log,
		cache: cache,
	}
}

// CacheAuthenticationSession takes the SRP object for the current authentication session and stores it somewhere.
// If it is able to store it successfully then a sessionId will be returned that can be used to retrieve the
// session from the cache.
func (s *srpCache) CacheAuthenticationSession(ctx context.Context, session *srp.SRP) (sessionId string, err error) {
	span := sentry.StartSpan(ctx, "CacheAuthenticationSession")
	defer span.Finish()

	{ // Generate a sessionId that we can use to store the authentication data.
		sessionIdBytes := make([]byte, 16)
		if _, err = rand.Read(sessionIdBytes); err != nil {
			return "", errors.Wrap(err, "failed to generate authentication session identifier")
		}
		// Then take the bytes and turn it into a more friendly identifier.
		sessionId = base32.StdEncoding.EncodeToString(sessionIdBytes)
	}

	// Encode the SRP data into a binary format.
	encodedSession, err := session.MarshalBinary()
	if err != nil {
		return "", errors.Wrap(err, "failed to encode authentication session")
	}

	// And yeet the data into our cache. If something goes wrong then return an error.
	return sessionId, errors.Wrap(
		s.cache.SetTTL(span.Context(), sessionId, encodedSession, 5*time.Minute),
		"failed to store authentication session",
	)
}

func (s *srpCache) LookupAuthenticationSession(ctx context.Context, sessionId string) (*srp.SRP, error) {
	span := sentry.StartSpan(ctx, "LookupAuthenticationSession")
	defer span.Finish()

	// Try to retrieve the data from the cache.
	encodedData, err := s.cache.Get(span.Context(), sessionId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve authentication session")
	}

	// If we didn't get an error then try to decode our byte array into the SRP object.
	session := new(srp.SRP)
	if err = session.UnmarshalBinary(encodedData); err != nil {
		return nil, errors.Wrap(err, "failed to read encoded authentication session")
	}

	// If the world hasn't fallen apart then we are probably fine and we can return this session to the caller.
	return session, nil
}
